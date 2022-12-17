```
FFS(1)                                                                                                                                                                                                                            User Manuals
NAME
       ffs - search for regex patterns in files

SYNOPSIS
       ffs [OPTION]... [ROOT]

DESCRIPTION
       ffs searches for regex patterns in files and prints the matching lines. The search can be 
       limited to specific file names or file contents using the -f and -s options, respectively, 
       or to hex-encoded lines using the -h option. The -b option can be used to include binary 
       files in the search. The search starts at the specified ROOT directory, or the current 
       directory if none is provided.

OPTIONS
       -f, --file
              regex pattern to match file names

       -s, --string
              regex pattern to match file string

       -h, --hex
              regex pattern to match hex-encoded lines

       -v, --verbose
              enable verbose mode

       -b, --binary
              include binary files in search

EXAMPLES
       Search for all files with the word "example" in their name under the current directory:
       
              ffs -f "example" .

       Search for all files with the word "password" in their contents under the current directory:
       
              ffs -s "password" .

       Search for all files with the hex encoded string "50 61 73 73 77 6f 72 64" in their contents
       under the current directory:
       
              ffs -h "50 61 73 73 77 6f 72 64" .

       Search for python files containing the string "import" and print the matching lines:

              ffs -f "\.py$" -s "import"

       Search for .c files containing the string "strcpy" and print the matching lines:

              ffs -f "\.c$" -s "strcpy"

       Search for files with names matching the pattern .log.\d, where \d is a digit, and for lines 
       that start with a datetime stamp in the range of 9:00:00 to 15:59:59. The ^ character indicates
       the start of the line, and the (09|10|11|12|13|14|15) pattern matches any of the given values. 
       
       The :[0-5][0-9]:[0-5][0-9] pattern matches any value in the range of 00:00 to 59:59 for the 
       minutes and seconds.

              ffs -f "\.log\.\d$" -s "^(09|10|11|12|13|14|15):[0-5][0-9]:[0-5][0-9]"

       Search to exclude the node_modules directory from the search:

              ffs -f '^(!.*node_modules).*$' -s 'react'

AUTHOR
       Eliot Alderson

COPYRIGHT
       Copyright (c) 2023 Eliot Alderson. All rights reserved.

LICENSE
       This program is free software: you can redistribute it and/or modify
       it under the terms of the GNU General Public License as published by
       the Free Software Foundation, either version 3 of the License, or
       (at your option) any later version.

       This program is distributed in the hope that it will be useful,
       but WITHOUT ANY WARRANTY; without even the implied warranty of
       MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
       GNU General Public License for more details.

       You should have received a copy of the GNU General Public License
       along with this program.  If not, see <https://www.gnu.org/licenses/>.

SEE ALSO
       grep(3)

FFS(1)
```
