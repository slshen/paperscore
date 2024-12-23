# paperscore

*paperscore* is a hybrid paper-computer softball scoring system that I've been using for the last couple of years to score amateur travel softball games.

The scoring notation is derived from [retrosheet](https://www.retrosheet.org) and is very similar to [Project Scoresheet](https://en.wikipedia.org/wiki/Baseball_scorekeeping#Project_Scoresheet).

Some features and notes:

* Generate box scores with `paperscore box`
* Edit game files with `paperscore ui`
* Export tournament, game, batting stats, and events to CSV format with `paperscore data-export`.  For my daughter's team I load these into a [hex.tech app](https://app.hex.tech/c3311da3-8517-4a59-a261-5fbb34686c1b/app/d06271cc-903f-4f37-8e55-9f141b1ea4f5/latest?).
* The scoring notation and software support entering "alternate plays", which is a play which in the scorers opinion should have occurred.  The software can compute the expected run cost of these errors and misplays.
* The software is incomplete, and undoubtedly has bugs.  Some of the code is experimental was just "left in" possibly to be completely later or more likely not at all.
* The error messages leave a lot to be desired.  I've found myself running the code in the debugger just to figure out what the actual error was.

## Sample Paper Score

Here's a sample score sheet:

![Sample Score](/doc/score.jpeg)

