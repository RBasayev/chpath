
A tool designed to change permissions or/and ownership along the path. Like this:
```
root@lh# chpath --mode og-xw /opt/java_apps/RAM-Optimizer/logs/npe.log
Skipping the first-level directory '/opt'.
changing mode 777 to 744 :    /opt/java_apps/RAM-Optimizer/logs/npe.log
changing mode 777 to 744 :    /opt/java_apps/RAM-Optimizer/logs
changing mode 777 to 744 :    /opt/java_apps/RAM-Optimizer
changing mode 777 to 744 :    /opt/java_apps

root@lh# chpath --reach /opt/java_apps/RAM-Optimizer/logs/npe.log
Setting permissions to reach
     /opt/java_apps/RAM-Optimizer/logs/npe.log
Skipping the first-level directory '/opt'.
changing mode 744 to 755 :    /opt/java_apps
changing mode 744 to 755 :    /opt/java_apps/RAM-Optimizer
changing mode 744 to 755 :    /opt/java_apps/RAM-Optimizer/logs
skipping, is 744 already :    /opt/java_apps/RAM-Optimizer/logs/npe.log
```

## Usage
```
       chpath <action flag> <flag argument> <action flag2> <flag argument> <path>
```

## Paths
```
  Path can be absolute or relative.

  Only ONE path will be processed.

  Before processing, the path is sanitized: symlinks will be resolved, . and .. will be
resolved etc. For example, the path '/a/b/c/S/../d/e' will become '/a/b/c/d/e'; the
directory S will not be processed, even if it exists.

  The top-level directory will always be skipped - as a safety measure. I.e., in the path:
     /nfs_exports/appOne/config/main.conf
     |           |             |<-target->|
     |<-skipped->|<------processed------->|
  
  The last element in the path is considered a TARGET.
```

## Action Flags and Their Arguments
```
  Every action flag may be specified only once. Strictly speaking, it can be specified multiple
times, but then only the last occurrence is considered.


-h
--help                       Display this Help and exit.


-v
--version                    Display the version and exit.


-r
--reach                      Adjust the mode to make the target "reachable", basically,
                             add 'x' along the path and add 'r' to the target (also 'x', if
                             target is a directory). This is done for owner, group and others.

                             When --reach is used, no arguments are expected. Also, no other
                             action flags are considered.

                             Example:
                             chpath --reach /opt/java_apps/RAM-Optimizer/logs/npe.log


-m
--mode                       Change the mode of all the elements in the path (after sanitizing).
                             Special bits (setuid etc.) are not (and will never be) implemented.

            ugoa+-rwxX       The general usage of mode arguments is the same as in 'chmod', with
                             very little differences. One of the differences - the order doesn't
                             matter - 'uwg+' will work the same as 'ug+w'. Why? I'm lazy.

                             Numeric notation of the permissions is not supported. Maybe one day.

                 u           Apply permissions to the file/directory owner.

                 g           Apply permissions to the file's/directory's group.

                 o           Apply permissions to "others" - neither owner, nor group members.

                 a           Apply permissions to owner, owner's group and for others. If 'a'
                             is specified, 'u', 'g' and 'o' are not considered.

               + / -         '+' adds the permissions, '-' removes them. If both are specified
                             (which makes no sense, but is not detected as an error) '+' wins.
                             When processing '+', chpath goes along the path from left to right,
                             and it goes from right to left, when processing '-'.

                              +    =1=>     =2=>          =3=> =4=>
                             /opt/java_apps/RAM-Optimizer/logs/npe.log
                              -        <=4=          <=3= <=2=    <=1=


                 r           Apply the READ permission.

                 w           Apply the WRITE permission.

                 x           Apply the EXECUTE permission.

                 X           Apply the EXECUTE permission only to directories (permission to
                             enter the directory).
                             N.B.: this is another difference with 'chmod' - 'chmod' will also
                                   apply the permission to files that have EXECUTE set for some
                                   other user; 'chpath' will not.
                             
                             Examples:
                             chpath --mode go-w /opt/java_apps/RAM-Optimizer/logs/npe.log
                             chpath --mode o+Xr /opt/java_apps/RAM-Optimizer/logs/npe.log


-o
--owner       username       Change the owner along the path.
               or uid        The new owner can be specified as username or as its numeric ID.
                             Specifying the user does not imply group change, it must be
                             specified explicitly with the next action flag.

-g
--group         name         Change the group along the path.
               or gid        The new group can be specified as name or as its numeric ID.

                             Owner and group change follows the path from left to right:

                                   =1=>     =2=>          =3=> =4=>
                             /opt/java_apps/RAM-Optimizer/logs/npe.log

                             Note: no shorthand syntax is implemented (yet?),
                                   i.e. something like "owner:group" in chown won't work.

                             Examples:
                             chpath --owner joe --group users /home/joe/.ssh/id_rsa
                             chpath --owner 0 /root/backups/june/backup.tgz


If multiple actions are specified, this is the order in which they are executed:
1. --help    (exits regardless of other arguments)
2. --version (exits regardless of other arguments)
3. --reach   (exits regardless of other arguments)
4. --mode
5. --owner and --group (always run together)


Combined actions example:
  chpath --owner joe --group staff --mode og-rwx /home/joe/.ssh/id_rsa

This will make sure that Joe owns the path to the SSH key and that no one else has access.

```

## Non-Zero Exit Codes

  3x - path inconsistencies
  
  4x - mode specification inconsistencies

  5x - user or group problems


## TODO

For parameters switch to https://github.com/jessevdk/go-flags
