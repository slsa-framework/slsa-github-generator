#!/bin/bash -eu
#
# Copyright 2009 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
################################################################################
#
# Outputs a header file comment, with the appropriate comments based on the
# language, as deduced from the extension of the file.
#
# Sample usage:
#   autogen.sh file.js
#   autogen.sh file.py

# If we're runnig via Bazel, find the source files via $TEST_SRCDIR;
# otherwise, default to dir of current file and search relative to that.
if [ -n "${TEST_SRCDIR:-}" ]; then
  declare -r SRCDIR="${TEST_SRCDIR}/${TEST_WORKSPACE}"
else
  # If the file pointing to this script is a symlink, we need to resolve it.
  # However, it may be pointing to another symlink, so we have to resolve the
  # entire chain to find the root, and find the licenses relative to the
  # directory that holds the file.
  #
  # Unfortunately, different tools are available on different OSes, and even
  # when the same tool is available, it works differently on different systems.
  if which realpath > /dev/null 2>&1 ; then
    declare -r SRCDIR="$(dirname $(realpath $0))"
  elif which greadlink > /dev/null 2>&1 ; then
    # On OS X, Homebrew provides `greadlink` which is GNU readlink that works as
    # it does on Linux. To get it, run: `brew install coreutils`.
    declare -r SRCDIR="$(dirname $(greadlink -f $0))"
  elif which readlink > /dev/null 2>&1 ; then
    # On Linux, `readlink -f` can be used to resolve symlinks to the final
    # destination. However, on OS X, `readlink` does not have a `-f` flag.
    if [[ "$(uname -s)" == "Darwin" ]]; then
      symlink_target="${0}"
      while [ -h "${symlink_target}" ] || [ -L "${symlink_target}" ]; do
        next_symlink_target="$(readlink "${symlink_target}" || true)"
        if [ -z "${next_symlink_target}" ] || [[ "${symlink_target}" == "${next_symlink_target}" ]]; then
          break
        fi
        symlink_target="${next_symlink_target}"
      done
      declare -r SRCDIR="$(dirname "${symlink_target}")"
      # Cleanup the variable namespace.
      unset symlink_target
      unset next_symlink_target
    else
      declare -r SRCDIR="$(dirname $(readlink -f $0))"
    fi
  else
    declare -r SRCDIR="$(dirname $0)"
  fi
fi

# Path to license file will be computed from LICENSE_NAME below.
LICENSE_NAME="apache"
LICENSE_FILE=""
COPYRIGHT_HOLDER="Google LLC"
YEAR="${YEAR:-$(date +%Y)}"
MODIFY_FILE_INPLACE=0
SILENT=0
ADD_CODE=1
ADD_RUNLINE=1
function printLicenseWithYear() {
  cat "${LICENSE_FILE}" \
    | sed "s/%YEAR%/${YEAR}/" \
    | sed "s/%COPYRIGHT_HOLDER%/${COPYRIGHT_HOLDER}/"
}

function printLicenseNonHashComment() {
  printLicenseWithYear | sed -E "s#^#$1 #;s/ +$//"
}

function printLicenseHashComment() {
  printLicenseWithYear | sed -E "s/^/# /;s/ +$//"
}

# Prepend piped-in text to a file in-place.
#
# Args:
#   $1: filename to modify
function prependToFileInPlace() {
  local -r file="${1}"

  local -r staging="$(mktemp /tmp/autogen-inplace.staging.XXXXXX)"
  # In case we exit due to an error early, clean up.
  # We also clean up explicitly below.
  trap "rm -f ${staging}" EXIT

  cat - "${file}" > "${staging}"
  cat "${staging}" > "${file}"

  # Future calls to `trap` overwrite previous ones, so make sure to delete our
  # temporary file.
  rm -f "${staging}"
}

SEPARATE_LICENSE_FROM_TODO="blank"
TOP_LEVEL_COMMENT=1
readonly TODO_COMMENT="TODO: High-level file comment."

function printFileCommentTemplate() {
  local comment=$1

  case "${SEPARATE_LICENSE_FROM_TODO}" in
    blank)
      echo
      ;;
    *)
      # Fit into 80 cols: repeat enough times, depending on our comment width.
      local repeat=$(perl -e "use POSIX; print floor(80 / length('${comment}'));")
      echo $comment
      perl -e "print \"$comment\" x $repeat . \"\n\""
      echo $comment
      ;;
  esac

  if [ ${TOP_LEVEL_COMMENT} -eq 1 ]; then
    echo "$comment ${TODO_COMMENT}"
    echo
  fi
}

function show_help() {
  cat << EOF
Syntax: $0 [options] <filename>

Options:
  --copyright [copyright-holder], -c [copyright-holder]
                           set copyright holder (default: "${COPYRIGHT_HOLDER}")

  --help, -h               show this help text

  --in-place, -i           modify the given file in-place (default: off, prints
                           output to stdout instead)

  --license [license], -l [license]
                           choose the license (default: "${LICENSE_NAME}")

  --separator              print a line of comment characters to separate
                           license from the top-level-comment (default: off,
                           prints a blank line instead)

  --silent, -s             silent: no error output for unknown file types; exits
                           with status code 0 (default: off)

  --top-level-comment, --tlc
  --no-top-level-comment, --no-tlc
                           add a top-level documentation comment after the
                           license header (default: enabled)

  --no-runline             Skip adding runline code
                           to the source files (default: false)

  --no-code                Skip adding bootstrap code
                           to the source files (default: false)

  --year [year], -y [year] choose the year (default: ${YEAR})

Licenses:
  apache2:      Apache 2.0 (alias: apache)
  bsd2:         BSD, 2-clause, aka Simplified/FreeBSD
  bsd3:         BSD, 3-clause, aka Revised/New/Modified
  bsd4:         BSD, 4-clause, aka Original
  gpl2:         GPL 2
  gpl3:         GPL 3
  lgpl2.1:      LGPL 2.1 (aliases: lgpl, lgpl2)
  mit:          MIT
  mpl2.0:       MPL 2.0 (aliases: mpl, mpl2)
EOF
}

while :; do
  case "${1:-}" in
    -c | --copyright)
      if [ -n "${2:-}" ]; then
        COPYRIGHT_HOLDER="${2}"
        shift
      else
        echo "Error: -c / --copyright requires an argument." >&2
      fi
      ;;

    -h | --help)
      show_help
      exit 0
      ;;

    -i | --in-place)
      MODIFY_FILE_INPLACE=1
      ;;

    -l | --license)
      if [ -n "${2:-}" ]; then
        LICENSE_NAME="${2}"
        shift
      else
        echo "Error: -l / --license requires an argument." >&2
      fi
      ;;

    -s | --silent)
      SILENT=1
      ;;

    --separator)
      SEPARATE_LICENSE_FROM_TODO="comment"
      ;;

    --top-level-comment | --tlc)
      TOP_LEVEL_COMMENT=1
      ;;

    --no-top-level-comment | --no-tlc)
      TOP_LEVEL_COMMENT=0
      ;;

    --no-code)
      ADD_CODE=0
      ;;

    --no-runline)
      ADD_RUNLINE=0
      ;;

    -y | --year)
      if [ -n "${2:-}" ]; then
        YEAR="${2}"
        shift
      else
        echo "Error: -y / --year requires an argument." >&2
      fi
      ;;

    --)
      shift
      break
      ;;

    -?*)
      echo "Invalid flag: ${1}." >&2
      show_help
      exit 1
      ;;

    *)
      # No more recognized options; switch to positional processing.
      break
      ;;
  esac

  shift
done

shift $((OPTIND - 1))

# Compute license file path given the license name.
case "${LICENSE_NAME}" in
  apache | apache2 | apache-2 | apache2.0 | apache-2.0)
    LICENSE_FILE="${SRCDIR}/licenses/apache-2.0.txt"
    ;;
  bsd2 | bsd-2)
    LICENSE_FILE="${SRCDIR}/licenses/bsd-2-clause.txt"
    ;;
  bsd3 | bsd-3)
    LICENSE_FILE="${SRCDIR}/licenses/bsd-3-clause.txt"
    ;;
  bsd4 | bsd-4)
    LICENSE_FILE="${SRCDIR}/licenses/bsd-4-clause.txt"
    ;;
  gpl2 | gpl-2)
    LICENSE_FILE="${SRCDIR}/licenses/gpl-2.txt"
    ;;
  gpl3 | gpl-3)
    LICENSE_FILE="${SRCDIR}/licenses/gpl-3.txt"
    ;;
  lgpl | lgpl2 | lgpl-2 | lgpl2.1 | lgpl-2.1)
    LICENSE_FILE="${SRCDIR}/licenses/lgpl-2.1.txt"
    ;;
  mit)
    LICENSE_FILE="${SRCDIR}/licenses/mit.txt"
    ;;
  mpl | mpl2 | mpl-2 | mpl2.0 | mpl-2.0)
    LICENSE_FILE="${SRCDIR}/licenses/mpl-2.0.txt"
    ;;
  *)
    echo "Invalid license selected: ${LICENSE_NAME}" >&2
    exit 1
esac

if [[ $# -eq 0 ]]; then
  show_help
  exit 1
fi

declare -r FILE="$1"

# Args:
#   $1: filename
function autogenForFile() {
  local file="$1"
  case "${file}" in

    *.bat)
      printLicenseNonHashComment "rem"
      echo
      if [ ${TOP_LEVEL_COMMENT} -eq 1 ]; then
        echo "rem ${TODO_COMMENT}"
        echo
      fi
      ;;

    # TODO(mbrukman): in some projects, *.h are actually C++ files where users
    # want to use C++ style "//" comments. How can we handle this flexibility?
    *.c | *.h | *.css | *.scss)
      echo "/*"
      printLicenseNonHashComment " *"
      echo " */"
      echo
      if [ ${TOP_LEVEL_COMMENT} -eq 1 ]; then
        echo "/* ${TODO_COMMENT} */"
        echo
      fi
      ;;

    *.cc | *.cpp | *.cs | *.dart | *.go | *.hh | *.hpp | *.java | *.js | *.jsx | *.ts | *.tsx | *.m | *.mm | *.proto | *.rs | *.scala | *.swift)
      printLicenseNonHashComment "//"
      printFileCommentTemplate "//"
      ;;

    *CMakeLists.txt | *.cmake)
      printLicenseHashComment
      printFileCommentTemplate "#"
      ;;

    *.el | .emacs | *.lisp)
      printLicenseNonHashComment ";;"
      printFileCommentTemplate ";;"
      ;;

    *.erl)
      printLicenseNonHashComment "%"
      printFileCommentTemplate "%"
      ;;

    *.hs)
      printLicenseNonHashComment "--"
      printFileCommentTemplate "--"
      ;;

    *.html | *.xml | *.vue)
      echo "<!--"
      printLicenseNonHashComment " "
      echo "-->"
      ;;

    *.jsonnet)
      printLicenseHashComment
      printFileCommentTemplate "#"
      ;;

    *.md | *.markdown)
      printLicenseWithYear
      ;;

    *.ml | *.sml)
      echo "(*"
      printLicenseNonHashComment " *"
      echo " *)"
      echo
      if [ ${TOP_LEVEL_COMMENT} -eq 1 ]; then
        echo "(* ${TODO_COMMENT} *)"
        echo
      fi
      ;;

    *.php)
      # We can't make PHP scripts locally executable with the #!/usr/bin/php line
      # because PHP comments only have meaning inside the <?php ... ?> which
      # means the first line cannot be simply #!/usr/bin/php to let the shell
      # know how to run these scripts.  Instead, we'll have to run them via
      # "php script.php" .
      #
      # Note: PHP accepts C, C++, and shell-style comments.
      echo "<?php"
      printLicenseNonHashComment "//"
      printFileCommentTemplate "//"
      echo
      # E_STRICT was added in PHP 5.0 and became included in E_ALL in PHP 6.0 .
      echo "error_reporting(E_ALL | E_STRICT);"
      echo "?>"
      ;;

    *.pl)
      if [ ${ADD_RUNLINE} -eq 1 ]; then
        echo "#!/usr/bin/perl"
        echo "#"
      fi
      printLicenseHashComment
      printFileCommentTemplate "#"
      if [ ${ADD_CODE} -eq 1 ]; then
        echo
        echo "use strict;"
      fi
      ;;

    test_*.py | *_test.py)
      if [ ${ADD_RUNLINE} -eq 1 ]; then
        echo "#!/usr/bin/python"
        echo "#"
      fi
      printLicenseHashComment
      echo
      if [ ${TOP_LEVEL_COMMENT} -eq 1 ]; then
        cat <<EOF
"""${TODO_COMMENT}"""

EOF
      fi
      if [ ${ADD_CODE} -eq 1 ]; then
        BASE_PY="${1/#test_/}"
        BASE_PY="${BASE_PY/_test/}"
        echo "import unittest"
        # Maybe import the package that this is intended to test.
        if [ -e "${BASE_PY}" ]; then
          echo "import ${BASE_PY/%.py/}"
        fi
        # Add basic bootstrap code.
        cat <<EOF


class FooTest(unittest.TestCase):

    def setUp(self):
        pass

    def tearDown(self):
        pass

    def testBar(self):
        pass


if __name__ == '__main__':
    unittest.main()
EOF
      fi
      ;;

    *.py)
      if [ ${ADD_RUNLINE} -eq 1 ]; then
        echo "#!/usr/bin/python"
        echo "#"
      fi
      printLicenseHashComment
      echo
      if [ ${TOP_LEVEL_COMMENT} -eq 1 ]; then
        cat <<EOF
"""${TODO_COMMENT}"""

EOF
      fi
      if [ ${ADD_CODE} -eq 1 ]; then
        cat <<EOF
import sys


def main(argv):
    pass


if __name__ == '__main__':
    main(sys.argv)
EOF
      fi
      ;;

    *.ipynb)
      printLicenseHashComment | $(dirname $0)/ipynb.py
      ;;

    *.rb)
      if [ ${ADD_RUNLINE} -eq 1 ]; then
        echo "#!/usr/bin/ruby"
        echo "#"
      fi
      printLicenseHashComment
      printFileCommentTemplate "#"
      ;;

    *.sh)
      if [ ${ADD_RUNLINE} -eq 1 ]; then
        echo "#!/bin/bash -eu"
        echo "#"
      fi
      printLicenseHashComment
      printFileCommentTemplate "#"
      ;;

    *.tcl)
      if [ ${ADD_RUNLINE} -eq 1 ]; then
        echo "#!/usr/bin/tclsh"
        echo "#"
      fi
      printLicenseHashComment
      printFileCommentTemplate "#"
      ;;

    *.tf | *.tfvars)
      printLicenseHashComment
      printFileCommentTemplate "#"
      ;;

    *.txt | README)
      printLicenseWithYear
      ;;

    *.vim | .vimrc | vimrc)
      printLicenseNonHashComment \"
      if [[ "${SEPARATE_LICENSE_FROM_TODO}" == "blank" ]]; then
        echo
      else
        # Handle the file header locally; hard to pass a double-quote to function
        # which wants to double-quote its arguments.
        echo "\""
        perl -e "print '\"' x 80 . \"\n\""
        echo "\""
      fi
      if [ ${TOP_LEVEL_COMMENT} -eq 1 ]; then
        echo "\" ${TODO_COMMENT}"
        echo
      fi
      ;;

    *.yaml | *.yml)
      printLicenseHashComment
      printFileCommentTemplate "#"
      ;;

    BUILD | Dockerfile | Makefile | Makefile.* | Rakefile | Vagrantfile)
      printLicenseHashComment
      printFileCommentTemplate "#"
      ;;

    *)
      if [ ${SILENT} -eq 0 ] ; then
        echo "File type not recognized: ${file}" >&2
        exit 1
      fi
      ;;

  esac
}

if [[ "${MODIFY_FILE_INPLACE}" -eq 1 ]]; then
  autogenForFile "${FILE}" | prependToFileInPlace "${FILE}"
  exit $?
else
  autogenForFile "${FILE}"
fi
