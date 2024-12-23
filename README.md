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
* There's no attempt to track players defensive positions.

## Sample Paper Score

Here's a sample score sheet:

![Sample Score](/doc/score.jpeg)

And here's it's translation into `.gm` format:

```
date: 6/9/2024
game: 2
visitorid: pride-jf-16u
home: Mafia Bateman 09
timelimit: 95m
tournament: Zoom into June
---
visitorplays
pitching 5
1 26 CBSBFFS K
2 8 BCBBFS K
3 00 BFFX 53/G5
score 0
4 2 FBBCBS K
5 17 BCFBBX 9/F89
6 6 X 5/P5
score 0
7 9 CBX S6/P5D
8 3 BX 3/P3D
9 12 CX 23/BG2/SH 1-2
10 21 BFX D7/L7 2-H
11 7 BLFB WP 2-3
  ... S K
score 1
12 11 FBBX S8/G4
13 18 X 8/F8
14 26 FX 53/G5 1-2
15 8 X 63/G6
16 00 CBCX E4/P4 B-1
17 2 FBBBX 7/F7
18 17 X 9/F9
19 6 SCFX 6/P6
20 9 SBBBFB W
21 3 FFBB NP conf
  ... FFBFS K
22 12 CBX 8/F8
23 21 BSFC K
score 1
24 7 BX S6/G6
25 11 X H/F8 1-H
26 18 CSFFX H/L8 conf
pitching 24
27 26 FBSBC K
28 8 CX 43/G4
29 00 CX 7/L7
final 4
homeplays
pitching 00
1 33 BBBCFS K
2 5 BFBBX 4/P3D
3 35 CBSFFT K
4 2 CSFBFS K
5 8 FBBX 53/G5
6 24 FTS K
7 87 FBBX 6/P46
8 13 SX S9/G9
9 22 SFBFS K
10 27 SCC K
11 10 BFBCFX S8/G4
12 33 MB CS2(26)
  ... X 7/F7
13 5 BSBX 7/L7
14 35 SCFBS K
15 2 BX 13/G1
16 8 SFS K
score 0
17 24 BX 43/G4
18 87 BFCBX E6/G6
  alt 63/G6
19 13 X D9/L89 1-3
20 22 X 3/BP3
21 27 BSSC K
score 0
22 10 BSBFC K
23 33 FSBS K
24 5 CBBX 7/F7
final 0
```

## Alternate Plays

From the game above we have the lines:

```
18 87 BFCBX E6/G6
  alt 63/G6
```

This translates as in the 18th PA, #87 reached base on an error by the shortstop on a ground ball.

The alternate reality (aka what they scorer think should have happened) is a 63 putout.

If we use the Division 3 run expectancy matrix, then this error cost 0.56 runs:

```
 In               Play              RCost        Comment       
B6.1 E6/G6                            0.56                     
```

Since recording an alternate play is completely up to the scorer, it can include plays that aren't by-the-book errors (ball drops at the feet of 3 players who are all staring at each other) or even include running plays.
