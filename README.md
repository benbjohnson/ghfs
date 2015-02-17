ghfs [![godoc](https://godoc.org/github.com/benbjohnson/ghfs?status.png)](https://godoc.org/github.com/benbjohnson/ghfs) ![Version](http://img.shields.io/badge/alpha-red.png)
====

The GitHub Filesystem (GHFS) is a user space filesystem that overlays the
GitHub API. It allows you to access repositories and files using standard
Unix commands such as `ls` and `cat`.


## Install

To use ghfs, you'll need to install [Go][go]. If you're running OS X then you'll
also need to install [MacFUSE][macfuse]. Then you can install ghfs by running:

```sh
$ go get github.com/benbjohnson/ghfs/...
```

This will install `ghfs` into your `$GOBIN` directory. Next you'll need to
create a directory and use `ghfs` to mount GitHub:

```sh
$ mkdir ~/github
$ ghfs ~/github &
```

Now you can read data from the GitHub API via the `~/github` directory.

[go]: https://golang.org
[macfuse]: https://osxfuse.github.io


## Usage

GHFS uses GitHub URL conventions for pathing. For example, to go to a user
you can `cd` using their username:

```sh
$ cd ~/github/boltdb
```

To go to a repository, you can use the username and repository name:

```sh
$ cd ~/github/boltdb/bolt
```

Once you're in a repository, you can list files using `ls` and you can print
out file contents using the `cat` tool.

```sh
bolt $ cat LICENSE
The MIT License (MIT)

Copyright (c) 2013 Ben Johnson

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
...
```


