package syncer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
	"github.com/mergestat/fuse/internal/db"
	uuid "github.com/satori/go.uuid"
)

const (
	reviewsFullSyncDays = 90 // 90 days

	selectGitHubPRReviews = `SELECT github_prs.number AS pr_number, github_pr_reviews.* FROM github_prs(?), github_pr_reviews(?, github_prs.number) ORDER BY github_pr_reviews.created_at DESC`
)

type githubPRReview struct {
	PRNumber                  *int       `db:"pr_number"`
	ID                        *string    `db:"id"`
	AuthorLogin               *string    `db:"author_login"`
	AuthorURL                 *string    `db:"author_url"`
	AuthorAssociation         *string    `db:"author_association"`
	AuthorCanPushToRepository *bool      `db:"author_can_push_to_repository"`
	Body                      *string    `db:"body"`
	CommentCount              *int       `db:"comment_count"`
	CreatedAt                 *time.Time `db:"created_at"`
	CreatedViaEmail           *bool      `db:"created_via_email"`
	EditorLogin               *string    `db:"editor_login"`
	LastEditedAt              *time.Time `db:"last_edited_at"`
	PublishedAt               *time.Time `db:"published_at"`
	State                     *string    `db:"state"`
	SubmittedAt               *time.Time `db:"submitted_at"`
	UpdatedAt                 *time.Time `db:"updated_at"`
}

// sendBatchGitHubPRReviews uses the pg COPY protocol to send a batch of GitHub pr reviews
func (w *worker) sendBatchGitHubPRReviews(ctx context.Context, tx pgx.Tx, repo uuid.UUID, batch []*githubPRReview) error {
	cols := []string{
		"repo_id",
		"pr_number",
		"id",
		"author_login",
		"author_url",
		"author_association",
		"author_can_push_to_repository",
		"body",
		"comment_count",
		"created_at",
		"created_via_email",
		"editor_login",
		"last_edited_at",
		"published_at",
		"state",
		"submitted_at",
		"updated_at",
	}

	inputs := make([][]interface{}, 0, len(batch))
	for _, review := range batch {
		input := []interface{}{
			repo,
			review.PRNumber,
			review.ID,
			review.AuthorLogin,
			review.AuthorURL,
			review.AuthorAssociation,
			review.AuthorCanPushToRepository,
			review.Body,
			review.CommentCount,
			review.CreatedAt,
			review.CreatedViaEmail,
			review.EditorLogin,
			review.LastEditedAt,
			review.PublishedAt,
			review.State,
			review.SubmittedAt,
			review.UpdatedAt,
		}
		inputs = append(inputs, input)
	}

	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"github_pull_request_reviews"}, cols, pgx.CopyFromRows(inputs)); err != nil {
		return fmt.Errorf("tx copy from: %w", err)
	}
	return nil
}

func (w *worker) handleGitHubPRReviews(ctx context.Context, j *db.DequeueSyncJobRow) error {
	l := w.loggerForJob(j)

	if err := w.sendBatchLogMessages(ctx, []*syncLog{
		{
			Type:            SyncLogTypeInfo,
			RepoSyncQueueID: j.ID,
			Message:         "starting to execute GitHub PR reviews lookup query",
		},
	}); err != nil {
		return err
	}

	id, err := uuid.FromString(j.RepoID.String())
	if err != nil {
		return fmt.Errorf("parse uuid: %w", err)
	}
	u, err := url.Parse(j.Repo)
	if err != nil {
		return fmt.Errorf("url parse: %w", err)
	}

	components := strings.Split(u.Path, "/")
	repoOwner := components[1]
	repoName := components[2]
	repoFullName := fmt.Sprintf("%s/%s", repoOwner, repoName)

	var tx pgx.Tx
	if tx, err = w.pool.BeginTx(ctx, pgx.TxOptions{}); err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				w.logger.Err(err).Msgf("rollback transaction: %v", err)
			}
		}
	}()

	// delete the recent rows within days for github_pull_request_reviews in PG
	sql := fmt.Sprintf("DELETE FROM github_pull_request_reviews WHERE repo_id = $1 and created_at > (now() - interval '%d day');", reviewsFullSyncDays)
	if res, err := tx.Exec(ctx, sql, j.RepoID.String()); err != nil {
		return fmt.Errorf("delete rows: %w", err)
	} else {
		l.Info().Msgf("deleted rows: %d", res.RowsAffected())
	}

	sql = "SELECT id FROM github_pull_request_reviews WHERE repo_id = $1 ORDER BY created_at DESC LIMIT 1"
	var reviewId string
	if err := tx.QueryRow(ctx, sql, id.String()).Scan(&reviewId); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("query row: %w", err)
		}
	}

	var rows *sqlx.Rows
	if rows, err = w.mergestat.QueryxContext(ctx, selectGitHubPRReviews, repoFullName, repoFullName); err != nil {
		return fmt.Errorf("mergestat query: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			w.logger.Err(err).Msgf("close rows: %v", err)
		}
	}()

	batchSize := 500
	batch := make([]*githubPRReview, 0, batchSize)
	totalCount := 0
	for rows.Next() {
		var r githubPRReview
		if err := rows.StructScan(&r); err != nil {
			return fmt.Errorf("row scan: %w", err)
		}
		if len(batch) >= batchSize {
			if err := w.sendBatchGitHubPRReviews(ctx, tx, id, batch); err != nil {
				return fmt.Errorf("batch insert pr reviews: %w", err)
			}
			batch = make([]*githubPRReview, 0, batchSize)
		} else {
			if *r.ID == reviewId {
				break
			}
			batch = append(batch, &r)
		}
		totalCount++
	}
	if len(batch) > 0 {
		if err := w.sendBatchGitHubPRReviews(ctx, tx, id, batch); err != nil {
			return fmt.Errorf("batch insert pr reviews: %w", err)
		}
	}

	l.Info().Msgf("retrieved PR reviews: %d", totalCount)

	if err := w.db.WithTx(tx).SetSyncJobStatus(ctx, db.SetSyncJobStatusParams{Status: "DONE", ID: j.ID}); err != nil {
		return fmt.Errorf("sync job done: %w", err)
	}

	return tx.Commit(ctx)
}
