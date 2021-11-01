# Architecture

See [this research paper](https://www.cs.usask.ca/~croy/papers/2013/AsaduzzamanICSM2013LineDraft.pdf) for details about the algorithm.

## Differences from the research paper

The paper describes the use of [simhash](https://en.wikipedia.org/wiki/SimHash) to improve performance. This implementation
does not perform this optimization because the performance seems "good enough".
