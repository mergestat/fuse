#!/bin/bash

set -e

db_add_repo(){
    local repo="$1"

    local is_github="false"

    if [[ "$repo" =~ ^https://github.com/ ]]; then
        is_github="true"
    fi

    echo "Creating repo: $repo" >&2

    psql -c "INSERT INTO repos (repo, is_github) VALUES ('$repo', $is_github)" $POSTGRES_CONNECTION

    local from_repo="FROM repos WHERE repo='$repo' AND is_github=$is_github"
    local repo_id="$(psql -Atqc "SELECT id $from_repo LIMIT 1" $POSTGRES_CONNECTION)"

    echo "Created Repo ($repo_id)" >&2
    psql -xc "SELECT * $from_repo" $POSTGRES_CONNECTION

    local sync_types="GIT_COMMITS GIT_COMMIT_STATS GIT_REFS GIT_FILES"

    if [ "$is_github" == "true" ]; then
        sync_types+=" GITHUB_REPO_METADATA GITHUB_REPO_ISSUES GITHUB_REPO_STARS GITHUB_REPO_PRS GITHUB_PR_COMMITS GITHUB_PR_REVIEWS"
    fi

    for sync_type in $sync_types; do
        echo "Adding repo sync for $sync_type"
        psql -c "INSERT INTO mergestat.repo_syncs (repo_id, sync_type) VALUES ('$repo_id', '$sync_type')" $POSTGRES_CONNECTION
        echo "Enqueuing repo sync for $sync_type"
        psql -c "INSERT INTO mergestat.repo_sync_queue (repo_sync_id, status) SELECT id, 'QUEUED' FROM mergestat.repo_syncs WHERE sync_type = '$sync_type'" $POSTGRES_CONNECTION
    done
}

if [ $# -ne 0 ]; then
    if [ "$1" == "add-repo" ]; then
        shift

        db_add_repo "$@"
    else
        exec "$@"
    fi

    exit $?
fi
