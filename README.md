[x] create scaffold for a golang project with gin
[x] if argv[1] == "run" then argv[2] is the directory with the data
[x] if argv[1] == "help" then print the help screen
[x] read authors.yaml - must not be empty
[x] read global configuration file
[x] use cmd line arguments to overwrite config values
[x] move everything to a github repository
[x] read files and .yaml metadata in each directory, descend recursively
    (do not read the files though)
[x] dynamically build all routes to the directories in "/content"
[x] Bug: do not create routes for subdirectories (e.g. /blog/2025)
[x] Bug: all URLs return the same content (post3.html)
[x] build page for a file on demand and in case it was not yet created
[x] store the generated file in a cache
[x] fetch the generated file from a cache

[x] assemble the layout (header and footer)
[x] return html instead of json, with the correct mimetype
[x] use picocss for a basic layout
[x] use go template functions to complete header and footer

[x] parse navigation.yaml into the Context structure
    [x] for each top level item in the navigation: create a route
    [x] repeat for each nested item (in the navigation)
    [x] for each *file* in the navigation directory: create a route
    [x] for each *subdir* in the navigation directory: create a route
    [x] and store the meta-information in the navigation directory
    [x] "/" in a directory redirects to the index page
    [x] enforce absolute paths (urls and locals) in navigation.yaml
[x] DataTree is then no longer required
[x] Hide file extension in the route (e.g. /index.md -> /index)

[x] add favicon to site configuration (under "branding")
[x] add route for static files (router.Static("/static", "./local-assets")
[x] review templating - do we have enough data for a header/footer?
    copy Navigation, (parts of)Config, Branding, Users
    also copy relevant stuff from Directory, File (merge them - settings in
    File have higher priority than those in Directory)
[x] move custom css template (site.css) to Branding, use it in header.html
    adjust header.html, footer.html
[x] remove Directory and File structure in Navigation; Directory is only
    required from the File struct itself (as a reference), but it is
    not required as a standalone object
[x] File structure is only required in the LookupIndex
[x] make File.CachedContent a byte array, not a string
[x] if the data tree changes, then all dependent files need to be regenerated
    [x] $site/layout: invalidate everything (i.e. delete CachedContent of all
        text/html files)
    [x] directory: everything including/below that directory is
        invalidated
    [x] directory metadata: everything including/below that directory is
        invalidated
    [x] file metadata: file is invalidated
    [x] file: file is recreated

[x] Create a default template for a minimalistic digital business card
    [x] Create a new layout for the new page
    [x] break it up into multiple html files
    [x] Move inline css to separate file
    [x] update config, navigation.yaml
    [x] update the main page (CV) if necessary
    [x] display date of last update (of the current page) in the footer
    [x] add a favicon (default symbol: • or ·)
    [x] use templating to add title, description
    [x] Check css and html with a linter, and format them properly
    [x] use templating to add navigation links
    [x] cmd line args ("create business-card-01 --out=directory") then copy this
        template to a new (clean!) subdirectory!

[x] Build automated tests
    [x] Support config option "dump template --out=directory"
    [x] Dump the whole context (including configuration, navigation etc)
        to a file ($out/context.json)
    [x] Also include the LookupIndex in the dump
    [x] Create html files for a whole site, dump them to a temporary directory
    [x] Then compare the output against .golden files
    [x] Look for a better library to parse command line args
        https://github.com/jessevdk/go-flags

[x] Migrate crupp.de to the new solution
    [x] Add a command line option 'version' to print the version
    [x] Reduce width of the layout - it is too wide right now
    [x] Improve readability of the projects page
    [x] Automate the deployment, e.g. in a docker container
        [x] Try without docker first, just by copying and running the binary
        [x] Systemd sample file: https://rootknecht.net/knowledge/linux/systemd/#simple-generic-service-file
    [x] Set up monitoring (uptimerobot.com)

!!!
!!! When running serve in ~/prj/miniblog, it will also react to changes
!!! in README.md (which is outside of the file tree)

[ ] Use case: host technical documentation
    [x] Rename impl to core
    [x] Move command line option handlers to cmd (help, version, run, dump)
    [x] Rename "dump" command line option to "static" (including Makefile!)
    [x] Move file generation logic to new file (content.go)
    [ ] Add metadata option to ignore header/footer 
        [ ] Add this to the tests
    [ ] Introduce a plugin mechanism
        [x] ContentTypePlugins depend on content type and file extension, and transform
        a whole file
        [x] Rewrite current logic as a new plugin
        [ ] The plugins decide about the mimetype
        [ ] The plugins decide whether header/footer is included (false for text/html)
        -> this is stored in the metadata, and evaluated in the router
    [ ] Support inline yaml for metadata (not in a separate file!)
        [ ] Should we still support the old file format? - yes!
        [ ] Add this to the tests
    [ ] Raw text is used as is, without header/footer
        [ ] This is a new plugin (plugins/contenttype/text.go)
    [ ] Support markdown templating and formatting
    use github.com/yuin/goldmark
        [ ] This is a new plugin (plugins/contenttype/markdown.go)
    [ ] Source code is formatted in a different style, with syntax highlighting
    [ ] Self-host the documentation of what we have built so far
    [ ] Support documentation for different versions, e.g. of an API
    [ ] Add search functionality, with key words and full text
    [ ] Show date of last (file) update In the footer of each page
    [ ] Expand the test suite

[ ] Do a major round of refactoring
    [ ] Review everything - is it idiomatic golang code?
    [ ] Check the wordpress interface for plugins - did we miss anything?

[ ] Use case: personal blog
    [ ] Default index page shows all posts (configurable!)
        [ ] This is a new plugin type (SnippetPlugin?)
    [ ] Site configuration specifies pagination etc
    [ ] Add RSS/Atom functionality
        [ ] This is a new plugin (plugins/contenttype/rss.go)
        [ ] This is a new plugin (plugins/contenttype/atom.go)
    [ ] Recreate auto-generated blog file if a post has been changed or
        added/removed
    [ ] Filter robots and spammers (see https://lambdacreate.com/posts/68)

[ ] if a file or directory was added then add the route
[ ] if a file or directory was removed then drop the route
