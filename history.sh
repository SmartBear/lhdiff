#!/usr/bin/env bash
#
# This script builds a sqlite database of lines and their unique ids.
#
set -e
IFS=$'\n'
NL=$'\n'

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

    if [[ $newFile != "cmd/lhdiff/main.go" ]]; then
      continue
    fi

    if [[ ${doLhdiff} = true ]]; then
      echo "----- ${sha}"
      pairs=$(lhdiff --omit <(git show $sha^1:$oldFile) <(git show $sha:$newFile))
#      echo "$pairs"

      lineIdsToRemove="'dummy'"
      insertStatements=""

      for pair in $pairs; do
        echo "$pair"
        oldLineNumber=$(echo $pair | cut -f1 -d,)
        newLineNumber=$(echo $pair | cut -f2 -d,)
        oldLineId=''
        newLineId=''

        # echo "${oldLineNumber} - ${newLineNumber}"

        if [[ ${oldLineNumber} != "_" ]]; then
          oldLineId=$(sqlite3 lines.db "SELECT line_id FROM lines WHERE file='${oldFile}' AND line_number=${oldLineNumber};")
          if [[ ! -z ${oldLineId} ]]; then
            lineIdsToRemove="${lineIdsToRemove},${NL} '${oldLineId}'"
          fi
        fi

        if [[ ${newLineNumber} != "_" ]]; then
          newLineId=$(sqlite3 lines.db "SELECT line_id FROM lines WHERE file='${newFile}' AND line_number=${newLineNumber};")
          if [[ ! -z ${newLineId} ]]; then
            lineIdsToRemove="${lineIdsToRemove},${NL} '${newLineId}', '__'"
#          else
#            newLineId="${sha}:${newFile}:${newLineNumber}"
          fi
        fi

        if [[ ${oldLineNumber} != "_" && ${newLineNumber}  = '_' ]]; then
          # Deleted
          type="D"
        else
          # Modified or Added
          if [[ ! -z ${oldLineId} ]]; then
            sql="INSERT INTO lines (line_id, file, line_number, sha) VALUES ('${oldLineId}', '${newFile}', ${newLineNumber}, '${sha}');;"
            insertStatements="${insertStatements}${NL}${sql}"
          else
            newLineId="${sha}:${newFile}:${newLineNumber}"
            sql="INSERT INTO lines (line_id, file, line_number, sha) VALUES ('${newLineId}', '${newFile}', ${newLineNumber}, '${sha}');"
            insertStatements="${insertStatements}${NL}${sql}"
          fi

          if [[ ${oldLineNumber} != "_" && ${newLineNumber} != "_" ]]; then
            # Modified
            type="M"
          elif [[ -z ${oldLineNumber} && ${newLineNumber} != "_" ]]; then
            # Added
            type="A"
          fi
        fi
      done

      if [[ ${lineIdsToRemove} != "'dummy'" ]]; then
        sqlite3 -echo lines.db "DELETE FROM lines where line_id IN (${lineIdsToRemove});"
      fi
      if [[ ${insertStatements} != '' ]]; then
        echo "${insertStatements}"
        echo "----"
        sqlite3 -echo lines.db "${insertStatements}"
      fi
    fi
  done
done
