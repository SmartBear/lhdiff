#!/usr/bin/env bash
#
# This script builds a sqlite database of lines and their unique ids.
# TODO: Query the database for the last commit, and start with that commit.
# This way the script can be run whenever there are new commits, and the
# database will just be updated.
#
# TODO: build a second table that stores the modified line ids for each commit:
# - sha
# - id
# - type = A(dded), M(odified), D(eleted)
#
set -e
IFS=$'\n'

sqlite3 lines.db <<EOF
CREATE TABLE IF NOT EXISTS LINES (
  id TEXT,
  file TEXT,
  line INTEGER,
  sha TEXT,
  PRIMARY KEY ( file, line ),
  CONSTRAINT unique_id UNIQUE (id)
);
EOF

# Loop over all commits, starting with the oldest revision
for sha in $(git log --reverse --pretty=format:"%h"); do
  # Loop over all files in the commit. We're interested in the files with one of the following formats:
  #   M    file                  # Modified file.
  #   Rn   oldfile   newfile     # Renamed file. The n represents the similarity. 100 means no content change.
  for diffLine in $(git --no-pager diff --name-status --find-renames=50% $sha^1 $sha 2>/dev/null); do
    type=$(echo $diffLine | awk '{ print substr($1,1,1) }')
    left=$(echo $diffLine | awk '{ print $2 }')
    right=$(echo $diffLine | awk '{ print $3 }')

    doLhdiff=false

    if [ $type == "M" ]; then
      right=$left
      doLhdiff=true
    fi
    if [ $type == "R" ]; then
      doLhdiff=true
    fi
    # TODO: handle deleted files - we should delete all relevant rows


    if [ $doLhdiff = true ]; then
      pairs=$(lhdiff <(git show $sha^1:$left) <(git show $sha:$right))
      for pair in $pairs; do
        old=$(echo $pair | cut -f1 -d,)
        new=$(echo $pair | cut -f2 -d,)
        id=$(sqlite3 lines.db "SELECT id FROM LINES WHERE file='$left' AND line=$old;")
        echo "${pair} => ${id}"
        if [[ -z $id ]]; then
          if [[ $new != "_" ]]; then
            # A new line that's never been seen before
            id="$sha:$left:$new"
            # There might already be a row with this line - delete it
            sqlite3 -echo lines.db "DELETE FROM LINES WHERE file='$right' AND line=$new;"
            sqlite3 -echo lines.db "INSERT INTO LINES (id, file, line, sha) values ('$id', '$right', $new, '$sha');"
          fi
        elif [[ $new != "_" ]]; then
          if [[ $left != $right || $old != $new ]]; then
            # An existing line that's been updated.
            # We have to be careful about not violiting the unique constraint here (creating duplicates).
            # If there is an exising row with the same line as $new, we first delete any row with the same id, then
            # update the id of that row.
            # Otherwise, we update the old row.
            idSameNew=$(sqlite3 lines.db "SELECT id FROM LINES WHERE file='$right' AND line=$new;")
            if [[ ! -z $idSameNew ]]; then
              sqlite3 -echo lines.db "DELETE FROM LINES WHERE id='$id';"
              sqlite3 -echo lines.db "UPDATE LINES SET id='$id', sha='$sha' WHERE id='$idSameNew';"
            else
              sqlite3 -echo lines.db "UPDATE LINES SET file='$right', line=$new, sha='$sha' WHERE id='$id';"
            fi
          fi
        else
          # A line that's been deleted
          sqlite3 -echo lines.db "DELETE FROM LINES WHERE id='$id';"
        fi
      done
    fi
  done
done
