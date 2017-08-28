# update\_cdstats

keep a small textfile with some stats about recently visited directories

It's just the go part. It's the same as this bash function:

```bash
update_cdstat() {
  local max_stat=100
  [ "$1" == "" ] && return
  local newpath=$1
  # count date_sec path
  local found=0
  local now=`date +%s`
  { while read count date path ; do
    [ "$path" == "$newpath" ] && { ((count++)) ; date=$now ; found=1 ; }
    echo "$count $date $path"
  done < ~/.cdstat
  [ "$found" -eq 0 ] && echo "1 $now $newpath"
  } | sort -nr | tail -$max_stat > ~/.cdstat.tmp
  sync
  mv ~/.cdstat{.tmp,}
}
```

I have this in my .bashrc:

```bash
# use the go binary if available (10 times faster)
# go get github.com/nicollet/update_cdstat
update_cdstat_command=`which update_cdstat || echo update_cdstat`

cd() {
  [[ "${1}" == -* || "$1" == "" ]] && { builtin cd $* ; return $? ; }
  builtin cd $* && $update_cdstat_command `realpath $1`
}
```

The go binary is 10 times faster, but as a user, there is no difference seen.

However, I may try to do more interesting things with that in the future.


