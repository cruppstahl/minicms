[ ] Build a graph of files and directories to simplify cache invalidation
    [x] Use a graph of File objects, and a hierarchy for Files/Directories
    [x] When reading the file system: always read file content (have to
        read the file anyway)
    [x] If a file body needs to be updated: delete the content
    [x] Handle new files, deleted files, new directories and deleted directories
    [x] Rebuild files if a dependency is modified
    [x] Also include File structures for the /layout path

    [x] The watcher goes through the tree and watches each directory
    [x] If a directory is modified then depending on the operation the tree
        is updated (e.g. parent directory is re-read from scratch)
        [x] ContentTypePlugins and DataPlugins will be invoked automatically
    [x] Review the plugins - they need to be threadsafe!
    [x] Stop supporting .yaml files with metadata for the files

    [x] Implement the existing plugins
        [x] builtin/text
        [x] builtin/markdown
        [x] builtin/html
        [x] builtin/search (as a dummy, for now)

    [x] Add helper functions for metadata ("GetBool" etc)
    [x] Only create the Search plugin on demand!
    [x] Store metadata as an opaque structure in File and Directory
    [x] Plugins need to assign metadata["ContentType"] (and others!)

    [x] Plugins need to skip frontmatter - but return it as metadata!
        [x] layout files do not have frontmatter!

    [x] When starting, print plugins and their priorities (ordered
        by priorities)!

    [x] Make sure that existing tests are running
        [x] /templates/business-01
        [x] /templates/documentation-01
        [x] /templates/current
        [x] `make test` needs to run successfully

    [x] FsWatcher needs to ignore temp files (e.g. index.html~)

    [x] Go back a step and review/draw happy/unhappy paths (use miro)
        [x] Plugins: Registration, helper functions
        [x] FileManager: Initialization
        [x] Router: Initialization
        [x] Router: Handling requests
        [x] FsWatcher: Initialization
        [x] FsWatcher: Registration, helper functions
        [x] FileManager: Initialize a single file (plugin flow)
        [x] FsWatcher triggered by a file update event

    [x] Use go's race detector to find multithreading issues
        go run -race main.go

    [x] Differentiate between "static" and "dump" - static just
        generates HTML files (w/o metadata etc)
        [x] "dump" generates HTML and metadata and configuration
        [x] 'make test' should run now

    [x] Perform more thorough reviews and create unittests
        [x] FileManager
        [x] Config
        [x] Plugin
        [x] Router
        [x] FsWatcher

    [x] make test should run unittests; also, it is currently flaky
        (IsActive flag is set randomly)
    [x] Add a MIT license file
    [x] Create a README with instructions on how to build & run everything

    [x] Use case: a new file is created
        - Make sure that a route is created!
    [x] Use case: a file is deleted
        - Make sure that the route is deleted!
    [ ] Use case: a new directory is created
        - Make sure that all routes are created!
        - Update the watcher as well
    [ ] Use case: a directory is deleted
        - Make sure that all routes are deleted!
        - Update the watcher as well

    [ ] Add more tests about file graph, hierarchy, metadata, plugins, router
        [ ] Systematically add/remove/update files and whole directories
        [ ] Update dependent files (layout/*)
        [ ] Add/remove metadata
        [ ] Make sure ModTime timestamp is updated
        [ ] Test against race conditions (how?)
        [ ] Make sure that caching is used correctly (i.e. files not updated
            unless it is necessary)

[ ] Add search functionality, with keywords and full text
    [x] Create a plugin interface with the following functions:
        [x] Initialize(params), Shutdown: for the plugin lifecycle
            Initialize returns more info, e.g. if it requires the full
            file body
        [x] AddFile: with every new file that was added
    [ ] Integrate bleve
        [x] Initialize the plugin, if enabled
        [x] Shutdown the plugin, if it exists
        [x] Feed it with all the files (and their cached content!)
        [ ] If the cache of a file is invalidated then delete the document
            from the search index
        [ ] Ignore files that have a ignore-for-search metadata flag
        [ ] If enabled, persist the index on disk

    [ ] Use metadata to decide whether a file should be added to
        the search index
    [ ] Create an endpoint to query the index (/q) if the plugin is enabled

[ ] Support custom 404 page (/content/404.\*), including metadata
    [ ] Add one to crupp.de
    [ ] Add this as a test

[ ] For runtime errors, show stack traces and debug info in the browser
    - only for DEBUG builds!
    - what should we do for non-debug builds? Silently ignore the errors
        and just log them? but then it would be tricky to detect them, so
        better scream loud!

[ ] Verify that the navigation can link to completely different directories
    (e.g. blog and documentation are maintained by different teams, and
    therefore stored in different repositories)
    -> is this a good idea? This opens the door to all kind of security
        issues
    -> better enforce that all paths are part of the /content root
        and that paths have to be RELATIVE (in the navigation and
        everywhere else!)
    -> Also, do NOT allow symbolic links! (for security reasons)

[ ] How about a new plugin to minify the html? (builtin/minifier)

