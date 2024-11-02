# paperscore

*paperscore* is a hybrid paper-computer softball scoring system that I've been using for the last couple of years to score amateur travel softball games.

Some features and notes:

* The software computes a variety of batting statistics that it can export in CSV format.  The CSV files can be imported into any of a number of other systems for additional analysis.  Personally I've been using [hex.tech](https://hex.tech).
* The scoring notation and software support entering "alternate plays", which is a play which in the scorers opinion should have occurred.  The software can compute the expected run cost of these errors and misplays.
* The scoring notation is derived from [retrosheet](https://www.retrosheet.org).
* The software is incomplete, and undoubtedly has bugs.  Some of the code is experimental was just "left in" possibly to be completely later or more likely not at all.
* The error messages leave a lot to be desired.  I've found myself running the code in the debugger just to figure out what the actual error was.
* There's a text UI editor for games that I find helpful.

## Sample Paper Score

Here's a sample score sheet:

![Sample Score](/doc/score.jpeg)

