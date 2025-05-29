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

!!!
!!! Bug: http://localhost:8080//blog/2025/01-post1 fails (two slashes!)

[ ] use case: digital business cards
    [x] parse navigation.yaml into the Context structure
    [ ] use it to build routes for the different directories and
        their index{.html|.md|.txt}
    [ ] as an alias, create an /index.html route as well
    [ ] site configuration has configuration about branding
    [ ] add routes for static files
    [ ] migrate crupp.de to the new solution
    [ ] build and deploy it (automate this step)

[ ] use case: host technical documentation
    [ ] self-host the documentation

[ ] use case: personal blog
    [ ] parse navigation.yaml and use it to build routes
    [ ] default index page shows all posts (configurable!)
    [ ] site configuration specifies pagination etc

[ ] if a file was updated, i.e. has a newer timestamp (and different checksum):
    rebuild it (but not more than once per minute)
[ ] if a file or directory was added then support the route
[ ] if a file or directory was removed then drop the route
[ ] make sure to also rebuild the generated pages etc
