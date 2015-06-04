#!/bin/bash

export PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin

# Enable alias expansion in command substitution
shopt -s expand_aliases

# Verify that we have gettext available
gettext "Bye." &> /dev/null

if [ $? -gt 0 ]; then
        alias ck_gettext="echo -e"
else
        alias ck_gettext="gettext -s --"
fi

## gettext echo with newline at the end
function echog() {
        local format="$1\n"
        shift
        printf -- "$(ck_gettext "$format")" "$@" >&2
}

## gettext echo with NO newline at the end
function   echogn() {
         local format="$1"
         shift
         printf -- "$(ck_gettext "$format")" "$@" >&2
}


function usage() {
        echo
        echog "Usage: buildrpm [options] [spec file]"
        echog "Options:"
        echog "-n | --name=<soft name>                          soft name then in dir"
        echog "-s | --spec=<filename>                           spec file path"
        exit 0
}

function help_notice() {
        echo
        echog "Use --help or -h to get more information"
        echo
        exit 1
}


CKNAME=`basename "$0"`
PARAMS=`getopt -a -n $CKNAME -o +n:s:r:hH -l name:,spec:,help,version,copyright,default -- "$@"`

[ $? -gt 0 ] && help_notice
eval set -- $PARAMS

[ "X$1" = 'X--' ] && usage

while [ "$1" != "--" ]; do
#       echo $1
        case "$1" in
                -h|-H|--help)
                        usage;;
                -n|--name)
                        shift
                        SOFTNAME=`eval echo $1`
                        ;;
                -s|--spec)
                        shift
                        SPECFILE=`eval echo $1`
                        ;;
        esac
        shift
done

if [ "X$SOFTNAME" = 'X' ];then
        echog "the --name option must given"
        exit 1
fi

if [ "X$SPECFILE" = 'X' ];then
        echog "the --spec option must given"
        exit 1
fi

# check build serial number
BUILDSN="$SOFTNAME/BUILDSN"
BUILD=`date '+%y%m%d000'`
if [ -e $BUILDSN ]
then
        BUILT=`cat $BUILDSN`
        [ $BUILD -gt $BUILT ] || BUILD=`expr $BUILT + 1`
else
        BUILT=$BUILD
fi
echo $BUILD > $BUILDSN
RELEASE="${BUILD}.el$(cat /etc/redhat-release|cut -d " " -f 3|cut -d "." -f 1)"
sed -i  "s/^Release:.*$/Release: "$RELEASE"/"  $SPECFILE


TMP_DIR="${SOFTNAME}/rpmbuild"
mkdir -p $TMP_DIR
mkdir -p $TMP_DIR/TMP
mkdir -p $TMP_DIR/BUILD
mkdir -p $TMP_DIR/RPMS
mkdir -p $TMP_DIR/SOURCES
mkdir -p $TMP_DIR/SRPMS
mkdir -p ${SOFTNAME}/RPMS



#if source point a http link then download the source to source dir
SOURCE=$(grep -oP 'Source:\s+(.*)' "$SPECFILE"| sed 's/Source://g')
if [ "X$SOURCE" != 'X' ];then
        cp $SOURCE -P $TMP_DIR/SOURCES/
fi

if [ -d "${SOFTNAME}/resource" ];then
        cp -r ${SOFTNAME}/resource/* $TMP_DIR/SOURCES/
fi

if [ -d "${SOFTNAME}/SRC" ];then
        cp -r ${SOFTNAME}/SRC/* $TMP_DIR/SOURCES/
fi
rpmbuild --define "_topdir $PWD/$TMP_DIR" \
 --define "_tmppath $PWD/$TMP_DIR/TMP" \
 --define "_rpmdir $PWD/$TMP_DIR/RPMS" \
 --define "_sourcedir $PWD/$TMP_DIR/SOURCES" \
 --define "_srcrpmdir $PWD/$TMP_DIR/SRPMS" \
 --define "_builddir  $PWD/$TMP_DIR/BUILD" \
 --nodeps -ba $SPECFILE

RPMBUILDERR=$?
if [ $RPMBUILDERR -ne 0 ];then
        echo "==========================================="
        echo "           rpmbuild exe failed             "
        echo "==========================================="
else
        find $PWD/$TMP_DIR/RPMS/ -name "*.rpm" -exec mv '{}' $PWD/${SOFTNAME}/RPMS/ \;
        rm -rf $TMP_DIR
        echo "==========================================="
        echo "           rpmbuild exe sucess             "
        echo "==========================================="
fi

exit $RPMBUILDERR
