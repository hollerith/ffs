```
FFS(1)                                                                                                                                                                                                                            User Manuals
NAME
       ffs - search for regex patterns in files

SYNOPSIS
       ffs [OPTION]... [ROOT]

DESCRIPTION
       ffs searches for regex patterns in files and prints the matching lines. The search can be
       limited to specific file names or file contents using the -f and -s options, respectively,
       or to hex-encoded lines using the -x option. The -b option can be used to exclude binary
       files from the search. The search starts at the specified ROOT directory, or the current
       directory if none is provided. If there is a .gitignore file in the directory then only
       files not ignored by git will be searched. No search criteria will list all files. The
       program can follow symlinks and recursion can be limited to a number of depths. File
       search patterns which look like common glob are converted to the equivalent regex. If
       the ROOT argument contains such a pattern then it acts a shorthand for a start directory
       and file match.

OPTIONS
       -f, --file=regex_pattern
              Search for files matching the given regex_pattern.

       -s, --string=regex_pattern
              Search for lines containing text matching the given regex_pattern.

       -x, --hex=regex_pattern
              Search for lines containing the hex-encoded bytes matching the given regex_pattern.

       -m, --meta=regex_pattern
              Search for metadata lines matching the given regex_pattern.

       -b, --binary
              Exclude binary files in the search. By default, binary files are included.

       -g, --gitignore
              When searching in directories containing a .gitignore file, ignore files and directories that
              would be ignored by git.

       -e, --errors
              Print any errors encountered during execution.

       -v, --verbose
              Print more information about what is happening and use a wide format file listing.

       -d, --depth=n
              Recurse at most n levels deep. The default is unlimited depth.

       -t, --trace
              Set debugging and tracing information during execution.

       -l, --links
              Follow symbolic links to directories.

       -h, --help
              Print usage information

EXAMPLES
       Search for all files with the word "example" in their name under the current directory:
              ffs -f "example" .

       Search for all files with the word "password" in their contents under the current directory:
              ffs -s "password" .

       Search for all PNG files under the current directory:
              ffs -m "image/png" .

       Search for all PNG files under the current directory:
              ffs -m "image/png" .

       Follow symlinks to search for all world executable files owned by root in /bin:
              ffs /bin -m "rwxr-xr-x.*0 - root" -l -v

       Find files with hex-encoded bytes "50 61 73 73 77 6f 72 64" in the current directory:
              ffs -x "50 61 73 73 77 6f 72 64" -d 0

       Search for python files containing the string "import" and print the matching lines:
              ffs -f "\.py$" -s "import"

       Search for .c files containing the string "strcpy" and print the matching lines:
              ffs -f "\.c$" -s "strcpy"

       Search for files with names matching the pattern .log.\d, where \d is a digit, and for lines that start
       with a datetime stamp in the range of 9:00:00 to 15:59:59. The ^ character indicates the start of the
       line, and the (09|10|11|12|13|14|15) pattern matches any of the given values. The :[0-5][0-9]:[0-5][0-9]
       pattern matches any value in the range of 00:00 to 59:59 for the minutes and seconds.
              ffs -f "\.log\.\d$" -s "^(09|10|11|12|13|14|15):[0-5][0-9]:[0-5][0-9]"

       Search only the node_modules directory from the search:
              ffs -f '^(.*node_modules).*$' -s 'react'

       List all files including not git version control in tests directory:
              ffs tests -g

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
