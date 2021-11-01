#!/usr/bin/env bash
IFS=$'\n'

# Loop over all commits, starting with the oldest revision
for sha in $(git log --reverse --pretty=format:"%h"); do
    # Loop over all files in the commit. We're interested in the files with one of the following formats:
    #   M    file                  # Modified file.
    #   Rn   oldfile   newfile     # Renamed file. The n represents the similarity. 100 means no content change.
    for diffLine in $(git --no-pager diff --name-status --find-renames=50% $sha^1 $sha 2> /dev/null); do
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

      if [ $doLhdiff = true ]; then
        for pair in $(lhdiff <( git show $sha^1:$left ) <( git show $sha:$right )); do
          echo "pair:$pair"
        done
      fi
    done
done

