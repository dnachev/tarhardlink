# tarhardlink
Tar file extractor, which uses hard links to save space if possible.

## Description
Given a series of Tar files, which are full snapshots of the same folder
in different point times (a backup using a Tar file is a good example),
`tarhardlink` can detect files, which haven't changed from the previous
archive and create a hard link instead of creating new copy of the file.
This way, extracting all Tar files will occupy the minimally required space
and allow to make it easy to work with the folder (no need to use `overlayfs`
or equivalent to provide a complete view of the folder).

This is very similar to the `rsync`'s `--link-dest` option, which creates
hardlinks to save space when creating backups.

## Comparison
In order to be as performant as possible, the command only checks the following
attributes of the file:
- size
- permissions
- modification time

Note: Files, which have been modified but for some reason have the same modification time
will still be hard linked.

## Gotchas
Because the files are hardlinked, they should be treated as read-only.
Any writes to a file in one of the folders, will change the file in all copies
which have hardlinked the file.

## Usage
After compiling, simply call:
```
tarhl -file <tar-file> -dest <target-folder> -base <base-folder>
```
where:
- `tar-file` is the Tar file to be extracted. If `-` is given, standard input
  will be used.
- `target-folder` is the name of the folder to extract the Tar file to. It will
  be automatically created if it doesn't exist.
- `base-folder` is the name of the folder, which will be used to check for creating
  hard links.

## Warranty
At the moment, this is a personal toy project and although it works for my use cases,
no guarantees are given it will work for yours. Be ready to debug and modify the utility
to make it run on your machine and fit your use case.

Also, I'm pretty sure there are potential security holes with malicious Tar files, e.g.
it doesn't validate file path names and it has the potential to escape the folder.