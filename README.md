xxo - Fast file opener
======================

I like fast "fuzzy" openers from TextMate, Eclipse, IntelliJ IDEA or plugins 
for Vim like CtrlP or Command-T. Those are very helpful tools.

But I want this in Bash! I very often need to open/edit files from about 20-30 
directories and I have to repeat myself writing those over and over again. 
Sure, I have some cd aliases, there are other options, but I like the "fuzzy" 
approach.

This is the purpose of this program. It is written in Go language because:

 * I want it _very_ fast while I do want to write it quickly.
 * I want to run it everywhere.

How to install
--------------

Currently there are no builds, you need to build manually:

    go install github.com/lzap/stringsim
    go clone github.com/lzap/xxo
    cd xxo
    go build

I will make this much more easier once I will have a time (autoloading of deps 
and standard binary go installation).

Current status
--------------

Working (fast) prototype :-)

Stay tuned

License
-------

GNU LGPL v3 or later: http://www.gnu.org/copyleft/lesser.html

vim: tw=79:fo+=w
