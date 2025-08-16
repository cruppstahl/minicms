# MiniCMS (Working Title)

A file-based content management system. 

## Table of contents

* [First Steps](#firststeps)
* [Usage](#usage)
* [Configuration](#configuration)
* [Themes](#themes)

## First Steps

To run minicms, i) download the source code, ii) build it, iii) run it.

```
git clone https://github.com/cruppstahl/minicms # i) download
cd minicms/cms
go build .                                      # ii) build
./cms run ../templates/business-card-01         # iii) run it
```

Congratulations! Your minicms server is now running on localhost, port 8080. If you [open a browser](https://localhost:8080) then you will see the page.

## Usage

Run `./cms -h` for a list of all supported command line options.

In most cases, though, you just want to run the server. Use
`./cms run <directory>`, where `<directory>` points to the directory where
your content is stored.

The content store requres a well-defined structure. Two example projects
are part of the repository:
 * `business-card-01` is an example for a digital business card
 * `documentation-01` is an example for technical documentation.

## Configuration

All configuration files are stored in the `<template>/config` directory.
In this directory you will find the following files:

  * `site.yaml` has site-wide configuration, e.g. port, but also the link
    to the favicon
  * `users.yaml` is a list of all authors - required, but not yet used
  * `navigation.yaml` stores the site's navigation

## Themes

Theme files are in `<template>/layout/header.html` and
`<template>/layout/footer.html`. Golang template language is supported.
(Documentation will be provided at a later stage. You will find a list of
template variables in `cms/plugins/helper.go:BuildTemplateVars`).

The CSS file is in `<template>/assets/site.css`.
