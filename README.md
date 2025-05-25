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
[ ] build (html) page for a file on demand and in case it was not yet created
[ ] store the generated file in a cache
[ ] fetch the generated file from a cache

!!!
!!! Bug: http://localhost:8080//blog/2025/01-post1 fails (two slashes!)

[ ] assemble the layout (header and footer)
[ ] use picocss for a basic layout
[ ] default index page shows all posts (configurable!)
[ ] "/" forwards to the index page

[ ] if a file was updated, i.e. has a newer timestamp (and different checksum):
    rebuild it (but not more than once per minute)
[ ] if a file or directory was added then support the route
[ ] if a file or directory was removed then drop the route

[ ] self-host the documentation
[ ] migrate crupp.de to the new solution
