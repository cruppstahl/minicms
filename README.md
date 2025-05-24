[x] create scaffold for a golang project with gin
[x] if argv[1] == "run" then argv[2] is the directory with the data
[x] if argv[1] == "help" then print the help screen

[x] read authors.yaml - must not be empty
[x] read global configuration file
[x] use cmd line arguments to overwrite config values
[ ] move everything to a github repository

[ ] build all routes to the directories in "/content"
[ ] also read the metadata.yaml files in each directory, and a list of files
    (do not read the files though)
[ ] build (html) page for a file on demand and in case it was not yet created
[ ] assemble the layout (header and footer)
[ ] use picocss for a basic layout
[ ] default index page shows all posts
[ ] "/" forwards to the index page

[ ] if a file was updated, i.e. has a newer timestamp (and different checksum):
    rebuild it
[ ] if a file or directory was added then support the route
[ ] if a file or directory was removed then drop the route
[ ] limit posts based on quota (number of files and of images)
[ ] self-host the documentation
[ ] migrate crupp.de to the new solution
