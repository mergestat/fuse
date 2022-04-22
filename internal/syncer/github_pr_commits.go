package syncer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/mergestat/fuse/internal/db"
	uuid "github.com/satori/go.uuid"
)

const selectGitHubPRCommits = `SELECT github_prs.number AS pr_number, github_pr_commits.* FROM github_prs(?), github_pr_commits(?, github_prs.number)`

type githubPRCommit struct {
	PRNumber       *int       `db:"pr_number"`
	Hash           *string    `db:"hash"`
	Message        *string    `db:"message"`
	AuthorName     *string    `db:"author_name"`
	AuthorEmail    *string    `db:"author_email"`
	AuthorWhen     *time.Time `db:"author_when"`
	CommitterName  *string    `db:"committer_name"`
	CommitterEmail *string    `db:"committer_email"`
	CommitterWhen  *time.Time `db:"committer_when"`
	Additions      *int       `db:"additions"`
	Deletions      *int       `db:"deletions"`
	ChangedFiles   *int       `db:"changed_files"`
	URL            *string    `db:"url"`
}

// sendBatchGitHubPRCommits uses the pg COPY protocol to send a batch of GitHub pr commits
func (w *worker) sendBatchGitHubPRCommits(ctx context.Context, tx pgx.Tx, j *db.DequeueSyncJobRow, batch []*githubPRCommit) error {
	cols := []string{
		"repo_id",
		"pr_number",
		"hash",
		"message",
		"author_name",
		"author_email",
		"author_when",
		"committer_name",
		"committer_email",
		"committer_when",
		"additions",
		"deletions",
		"changed_files",
		"url",
	}

	duplicated := make(map[string]bool)

	inputs := make([][]interface{}, 0, len(batch))
	for _, r := range batch {
		var repoID uuid.UUID
		var err error
		if repoID, err = uuid.FromString(j.RepoID.String()); err != nil {
			return fmt.Errorf("uuid: %w", err)
		}

		if duplicated[*r.Hash] {
			w.logger.Warn().Msgf("duplicated input: %s, %s", repoID.String(), *r.Hash)
			continue
		} else {
			duplicated[*r.Hash] = true
		}

		input := []interface{}{
			repoID,
			r.PRNumber,
			r.Hash,
			r.Message,
			r.AuthorName,
			r.AuthorEmail,
			r.AuthorWhen,
			r.CommitterName,
			r.CommitterEmail,
			r.CommitterWhen,
			r.Additions,
			r.Deletions,
			r.ChangedFiles,
			r.URL,
		}
		inputs = append(inputs, input)
	}

	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"github_pull_request_commits"}, cols, pgx.CopyFromRows(inputs)); err != nil {
		return fmt.Errorf("tx copy from: %w", err)
	}
	return nil
}

func (w *worker) handleGitHubPRCommits(ctx context.Context, j *db.DequeueSyncJobRow) error {
	l := w.loggerForJob(j)

	// indicate that we're starting query execution
	if err := w.sendBatchLogMessages(ctx, []*syncLog{
		{
			Type:            SyncLogTypeInfo,
			RepoSyncQueueID: j.ID,
			Message:         "starting to execute GitHub PR commits lookup query",
		},
	}); err != nil {
		return fmt.Errorf("log messages: %w", err)
	}

	var u *url.URL
	var err error
	if u, err = url.Parse(j.Repo); err != nil {
		return fmt.Errorf("url parse: %v", err)
	}

	components := strings.Split(u.Path, "/")
	repoOwner := components[1]
	repoName := components[2]
	repoFullName := fmt.Sprintf("%s/%s", repoOwner, repoName)

	commits := make([]*githubPRCommit, 0)
	if err := w.mergestat.SelectContext(ctx, &commits, selectGitHubPRCommits, repoFullName, repoFullName); err != nil {
		return fmt.Errorf("mergestat select: %w", err)
	}

	l.Info().Msgf("retrieved PR commits: %d", len(commits))

	var tx pgx.Tx
	if tx, err = w.pool.BeginTx(ctx, pgx.TxOptions{}); err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				w.logger.Err(err).Msgf("could not rollback transaction")
			}
		}
	}()

	if _, err := tx.Exec(ctx, "DELETE FROM github_pull_request_commits WHERE repo_id = $1;", j.RepoID.String()); err != nil {
		return fmt.Errorf("exec delete: %w", err)
	}

	if err := w.sendBatchGitHubPRCommits(ctx, tx, j, commits); err != nil {
		return fmt.Errorf("github pr commits: %w", err)
	}

	if err := w.db.WithTx(tx).SetSyncJobStatus(ctx, db.SetSyncJobStatusParams{Status: "DONE", ID: j.ID}); err != nil {
		return fmt.Errorf("update status done: %w", err)
	}

	return tx.Commit(ctx)
}
