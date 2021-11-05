#!/usr/bin/env bash
#
# This script builds a sqlite database of lines and their unique ids.
#
set -e
IFS=$'\n'

sqlite3 lines.db <<EOF
CREATE TABLE IF NOT EXISTS lines (
  line_id TEXT NOT NULL,
  file TEXT NOT NULL,
  line_number INTEGER NOT NULL,
  sha TEXT NOT NULL,
  UNIQUE(file, line_number),
  UNIQUE(line_id),
  CHECK (length(line_id) > 0)
);

CREATE TABLE IF NOT EXISTS changes (
  line_id TEXT NOT NULL,
  file TEXT NOT NULL,
  line_number INTEGER,
  sha TEXT NOT NULL,
  type TEXT NOT NULL,
  UNIQUE(line_id, sha),
  CHECK (length(line_id) > 0)
);
EOF

# Loop over all commits, starting with the oldest revision
for sha in $(git log --reverse --pretty=format:"%h"); do
  # Loop over all files in the commit. We're interested in the files with one of the following formats:
  #   M    file                  # Modified file.
  #   Rn   oldfile   newfile     # Renamed file. The n represents the similarity. 100 means no content change.
  for diffLine in $(git --no-pager diff --name-status --find-renames=50% $sha^1 $sha 2>/dev/null); do
    type=$(echo $diffLine | awk '{ print substr($1,1,1) }')
    oldFile=$(echo $diffLine | awk '{ print $2 }')
    newFile=$(echo $diffLine | awk '{ print $3 }')

    doLhdiff=false

    if [ $type == "M" ]; then
      newFile=$oldFile
      doLhdiff=true
    fi
    if [ $type == "R" ]; then
      doLhdiff=true
    fi
    # TODO: handle deleted files - we should delete all relevant rows

#    if [[ $newFile != "cmd/lhdiff/main.go" ]]; then
#      continue
#    fi

    if [[ ${doLhdiff} = true ]]; then
      pairs=$(lhdiff --omit <(git show $sha^1:$oldFile) <(git show $sha:$newFile))
      # echo "$pairs"
      for pair in $pairs; do
        # echo "$pair"
        oldLineNumber=$(echo $pair | cut -f1 -d,)
        newLineNumber=$(echo $pair | cut -f2 -d,)

        if [ ${oldLineNumber} == "_" ]; then
          # Added
          sqlite3 -echo lines.db "DELETE FROM lines WHERE file='${newFile}' AND line_number=${newLineNumber};"
          lineId=$(uuidgen)
          sqlite3 -echo lines.db "INSERT INTO lines (line_id, file, line_number, sha) VALUES ('${lineId}', '${newFile}', ${newLineNumber}, '${sha}');"

          lineNumber=${newLineNumber}
          type="A"
        elif [[ ${newLineNumber} != "_" ]]; then
          # Modified
          sqlite3 -echo lines.db "DELETE FROM lines WHERE file='${newFile}' AND line_number=${newLineNumber};"
          lineId=$(sqlite3 lines.db "SELECT line_id FROM lines WHERE file='${oldFile}' AND line_number=${oldLineNumber} AND sha IS NOT '${sha}';")
          if [[ -z ${lineId} ]]; then
            lineId=$(uuidgen)
          else
            sqlite3 -echo lines.db "DELETE FROM lines WHERE line_id='${lineId}';"
          fi
          sqlite3 -echo lines.db "INSERT INTO lines (line_id, file, line_number, sha) VALUES ('${lineId}', '${newFile}', ${newLineNumber}, '${sha}');"

          lineNumber=${newLineNumber}
          type="M"
        else
          # Deleted
          lineId=$(uuidgen)
          sqlite3 -echo lines.db "DELETE FROM lines WHERE file='${oldFile}' AND line_number=${oldLineNumber};"
          lineNumber=""
          type="D"
        fi

        sqlite3 -echo lines.db "INSERT INTO changes (line_id, file, line_number, sha, type) VALUES ('${lineId}', '${newFile}', ${lineNumber:-"NULL"}, '${sha}', '${type}');"
      done
    fi
  done
done
