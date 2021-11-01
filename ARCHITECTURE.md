# Architecture

The algorithm is as follows:

1. Trim whitespace from each line in both files
1. Perform a UNIX diff
1. Map unchanged lines
1. Analyzing the UNIX diff, let `leftLines` be the lines that are *deleted* from the left file, and `rightLines` the lines that are *added* to the right file.
1. Map each `leftLine` to a `rightLine` if their *distance* is smaller than a predefined threshold.
   The distance is a combination of [levenshtein distance](https://en.wikipedia.org/wiki/Levenshtein_distance) of the 
   two lines as well as the [cosine similarity](https://en.wikipedia.org/wiki/Cosine_similarity) of the context around each line.

See [this research paper](https://www.cs.usask.ca/~croy/papers/2013/AsaduzzamanICSM2013LineDraft.pdf) for details about the algorithm.

## Differences from the research paper

The paper describes the use of [simhash](https://en.wikipedia.org/wiki/SimHash) to improve performance. This implementation
does not perform this optimization because the performance seems "good enough".

The paper also describes an option to detect line spliting. This is not implemented.
