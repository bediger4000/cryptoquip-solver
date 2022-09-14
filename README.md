# Cryptoquip Solver

King Features has a daily cipher puzzle
called [Cryptoquip](https://weekly.kingfeatures.com/?team=games-and-puzzles)
that might run in your newspaper.

![Example Cryptoquip](cq.png)

People like this sort of thing enough that [books full](https://www.amazon.com/cryptoquip/s?k=cryptoquip)
of Cryptoquips exist.

The Cecil (Maryland) Whig newspaper seems to have a good [online archive of recent cryptoquips](https://www.cecildaily.com/diversions/cryptoquip/).
I find this strange.

There are several other solvers available that may suit your needs better.

https://www.quipqiup.com

https://rumkin.com/tools/cipher/cryptogram-solver/

## Method of Solving

Cryptoquips are plain alphabetic replacement ciphers.
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

It does cycle through the process more than once.
