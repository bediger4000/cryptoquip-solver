# Cryptoquip Solver

King Features has a daily cipher puzzle
called [Cryptoquip](https://weekly.kingfeatures.com/?team=games-and-puzzles)
that might run in your newspaper.

There's also [Celebrity Cipher](http://syndication.andrewsmcmeel.com/puzzles/celebritycipher)
from Andrews McMeel Syndication.
I think it's identical in all but name.
There's a weird [Celebrity Cipher answers site](https://celebritycipheranswers.com/).
It doesn't have any puzzles, just the answers, and it doesn't give a method of solution.
I suspect it only exists to collect email addresses.

![Example Cryptoquip](cq.png)

People like this sort of thing enough that [books full](https://www.amazon.com/cryptoquip/s?k=cryptoquip)
of Cryptoquips exist.

The Cecil (Maryland) Daily Whig newspaper seems to have a good [online archive of recent cryptoquips](https://www.cecildaily.com/diversions/cryptoquip/).
I find this strange.

There are other solvers available that may suit your needs better.

This looks interesting, but doesn't explain its algorithm: https://www.quipqiup.com

## Building and Running


```sh
$ cd $GOPATH/src
$ git clone https://github.com/bediger4000/cryptoquip-solver.git cryptoquip
$ cd cryptoquip
$ go build solver.go
```

You need to type in a Cryptoquip from a newspaper or something.
It should look a little like this:

```
x=g
tkdcfq pcdygjkv bec ucwyq zoyzkojvx dyks bjse k wyor qujxesur qzcjuyg tukwco
```

The "x=g" line is the hint the daily cryptoquip gives.

After that, you can run the program:

```sh
$ ./solver -p puzzle.in -v > puzzle.out
```

The default clear text dictionary is `/usr/share/dict/words`.
You may need to install that, it sometimes isn't in a distro's default packages.
I also find that `/usr/share/dict/words` has far too many not-really-words,
like lists of lower-case Roman numerals.
Apparently `words` intended use case is spell-checkers,
and folks don't like it flagging the lower-case Roman numerals used on forewards.
I've noticed that larger dictionaries don't give better results with my
[Jumble Solver](https://github.com/bediger4000/jumble-solver) either.
I don't know if this is a general, information-theoretic problem,
or if I've just stumbled across two peculiar cases.

The `-v` flag gives very verbose output that will help you see what the program does.

### Encoder

I also include a mono-alphabetic cipher construction and encoder.

```sh
$ go build encoder.go
$ ./encoder input.txt > ciphertext.out
```

The ciphertext output shows you the clear-to-cipher letter correspondence,
and helpfully puts in all possible "x=y" hints as comments.

You can construct your own Cryptoquips,
and have the fun of solving a puzzle that you already have an answer for.

### Find dictionary words by shape

```sh
$ go build findbykey.go
$ ./findbykey words footballer
```

See what dictionary words match (by "shape") specified words.

## The Program Will Have Problems

If the answer to the Cryptoquip includes a word that isn't in the dictionary,
my program will probably not find a solution.
It will find words of the same "shape" as the non-dictionary enciphered word,
but that may not include the correct clear text letter at in the set of clear text
letters corresponding to some enciphered letter.

This can show up as enciphered letters that don't get a single clear text letter
as a solution even after many cycles through the algorithm.

It can also show up as an enciphered letter that has at least 2 
"single" clear text letters when correlating regular expression matches.
See below.

Sometimmes the presence of a single word can cause the program problems.
Removing "xor" from my dictionary of clear text words let my program
solve all of some cryptoquips,
where it previously could not find the clear text solution to one cipher letter.

## Method of Solving

Cryptoquips are simple alphabetic replacement ciphers.
They're not quite as simple as a [Caesar cipher](https://en.wikipedia.org/wiki/Caesar_cipher)
in that if cipher letter A corresponds to clear text letter M,
a Cryptoquip's cipher letter B doesn't necessarily correspond to clear text letter N.
It's not just a rotation of the clear alphabet relative to the cipher alphabet.

The usual method of breaking these kinds of ciphers has you
[frequency analysis](https://www2.rivier.edu/faculty/vriabov/cs572aweb/Assignments/CrackingClassicCiphers.htm):
counting frequencies of cipher letters, and matching high-frequency-of-appearance cipher letters
to high-frequency-of-appearance letters in plain text.
There's not enough text in any day's Cryptoquip to allow frequency analysis to work.
Besides that, getting [frequency analysis](https://github.com/bediger4000/vigenere-ciphering-deciphering)
correct is bug-prone for some reason.
If you wanted to do successful frequency analysis on Cryptoquips,
you'd probably have to consider frequencies of bigrams
as well as single letters,
and also frequency of common short words.
It would be a hassle.

My contribution is to use the information in the arrangement of ciphertext into words.
A (whitespace separated) ciphered word has the same number of letters,
and arrangement of letters,
as its deciphered corresponding word.

The image of the puzzle above has the enciphered word OHDXLZSVPLYY.
The corresponding  deciphered word will have 12 characters,
and end with a pair of the identical characters.
I represented the "shape" and arrangement of letters with strings of digits.
Enciphered word OHDXLZSVPLYY has the "shape" "012345678499"

*Begin Cycle*

Looking through a dictionary of words (Linux `/usr/share/dict/words` with some deletions),
I find these possible cleartext words based on "shape" alone:

```
3 matches
        gracefulness
        gratefulness
        motherliness
Found letters for configuration 012345678499
Letters at 0: g m 
Letters at 1: o r 
Letters at 2: a t 
Letters at 3: c t h 
Letters at 4: e 
Letters at 5: f r 
Letters at 6: u l 
Letters at 7: l i 
Letters at 8: n 
Letters at 9: e 
Letters at 10: s 
Letters at 11: s 
```

By merely looking at clear text words of the same "shape",
collecting clear text letters at the same position inside words,
we can find clear text equivalents of 4 enciphered letters.

My program looks through the dictionary, grouping clear text words by "shape".
It finds all the possible letters at each letter's position inside any given word.
It collects all the possible letters for each enciphered letter.

For example,
the two enciphered words XQUPI and XQLUH
both have shapes "01234".
My dictionary of clear text words has 4240 words of shape "01234".
All 26 English letters could be the solution of 'X' at position 1.
For some word shapes, the information in length of words
and arrangement of letters doesn't help.

My progam also looks at correlations of enciphered letters between words.
For each enciphered word, it looks at the sets of clear text letters that correspond
to any enciphered letters between each word.

The 'X' in OHDXLZSVPLYY could be any of {c, t, h,}
Intersecting {c, t, h} with the 26 letters for 'X' in XQUPI and XQULH
only narrows clear text letter possibilities for 'X' to {c, t, h},
but in a lot of cases,
inserting all the possible clear text letters narrows the possibilities down considerable.

After my program has narrowed down the letters based on "shape" of all the enciphered words,
it creates regular expressions for the clear text words that could match each enciphered word.

Enciphered word XQLUH has a clear text solution that matches `'^[cht][a-ik-pr-z]e[acehilnoru-wy][or]$'`.
Enciphered word XQUPI has a clear text solution that matches `'^[cht][a-ik-pr-z][acehilnoru-wy]n[a-ik-pr-wyz]$`
My program tries all the 4240 5-letter clear text words that have shape "01245" against the regular expressions.
For the regular expression of XQLUH, it finds 2 clear text words, "clear" and "their".
My program can derive 2 enciphered letters' clear text, 'L' deciphers to 'e' and 'H' deciphers to 'r',
since those letters are the same in all same-shape clear text words that match the regular expression.

The regular expression derived from XQUPI matches 65 clear text words of the same shape,
but conveniently, enciphered letter P only as clear text letter 'n' in all 64 words.

My program also correlates clear text letters common to clear text words from 2 or more regular expressions.
It uses length, letter position, and common letters between words to narrow the clear text letters that could
possibly match an enciphered word.

My program creates a new dictionary of all possible letters that could match any given enciphered letter
by working through all the words that match the regular expressions.
This probably narrows the clear text letters somewhat, but throws away all of the single clear text letter
solutions found by either correlating letter positions inside words of the same "shape" or
single clear text letters in all same-shape-words that match the regular expressions.

On the next cycle, my program will use any clear text letters corresponding to enciphered letters
it has already found when it creates new regular expressions.
This causes regular expressions derived from the enciphered words to match fewer same-shape words from
the clear text dictionary.

It does cycle through the process more than once. _Go to Begin Cycle_

My program finds a solution to the puzzle above in 2 cycles:

```
tqlp ypdily qdfl xql glyuhl xr yqrt xqluh ohdxlzsvplyy gr crs xqupi xqlc tuvv oufl zdpoy
when snakes have the desire to show their gratefulness do you think they will give fangs

```

## Algorithm Failure Mode

This puzzle is an adversarial example for my algorithm:

```
peen loop over cool boot green bean steep peat
haaf lbbh bmat vbbl ibbs ntaaf iapf osaah haps
# clear   a b c e g l n o p r s t v
# cipher  p i v a n l f b h t o s m
```

The puzzle has only 3 word shapes: "0112, "0123", "01223".
By shape alone, my clear text dictionary has matches:

|0112|162|
|0123|2721|
|01223|433|

3316 shape-matches.

I looked up those sentences that have all 26 letters in the english
alphabet once, thinking that they would contain so few duplicate
letters that my solver would fail.

```
the quick brown fox jumps over the lazy red dog
sphinx of black quartz judge my vow
pack my box with five dozen liquor jugs
jackdaws love my big sphinx of quartz
the quick onyx goblin jumps over the lazy dwarf
#amazingly few discotheques provide jukeboxes
```

The first 5 lines amount to 210 bytes,
which is about the same size as many of the cryptoquips
my solver gets in just a few cycles.
My solver does not appear to ever solve the first 5 lines.
Adding the 6th, commented-out line, allows my solver to work.
In fact, any 4 of the first 5 lines plus the 6th line ("amazingly few ...")
are easily solvable.
I suspect this happens because the 6th line has duplicate 'o' and 'e' characters.
