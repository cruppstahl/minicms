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
[ ] if the data tree changes, then all dependent files need to be regenerated
	[ ] header or footer: everything will be recreated (i.e. delete
		CachedContent of all text/html files)
	[ ] directory metadata: everything below that directory is recreated
	[ ] file metadata: file is recreated
	[ ] file: file is recreated
	[ ] what about auto-generated blog files?

[ ] Create a default template for a minimalistic digital business card

!!!
!!! how would a minimalistic design look like for a digital business card with
!!! projects (e.g. open source projects), services (e.g. mentoring engineering
!!! managers, startup advisory), how to get in touch, an overview of the CV,
!!! /now, ...? 

(Intro text) I am [Head of Engineering/now] at EPI, where we are building a new pan-European payments scheme called Wero[link]. My background is in software engineering [CV], mostly with C/C++. I have written a variety of open source projects, most of them outdated by now[projects]. I have published research work with [Daniel Lemire] on compression algorithms in databases. If you are an engineering manager, I am available for mentoring. If you work for a startup, I am available as a startup advisor.

(Then add how to get in touch - email, linkedin)

    [ ] update config, navigation.yaml
    [ ] update the main page (CV) if necessary
    [ ] add a /now page
    [ ] add a favicon (default symbol: • or ·)
    [ ] display date of last update (of the current page) in the footer
    [ ] footer: also add a "built with..." and a link to the github repository
    [ ] add a pdf with a CV
    [ ] use templating to add links etc, instead of hardcoding them
    [ ] cmd line args ("create --template=business-card-01") then copy this
        template to a new (clean!) subdirectory!
    [ ] migrate crupp.de to the new solution
    	[ ] automate the deployment, e.g. in a docker container
    	[ ] set up monitoring

[ ] use case: host technical documentation
    [ ] self-host the documentation of what we have built so far

[ ] use case: personal blog
    [ ] parse navigation.yaml and use it to build routes
    [ ] default index page shows all posts (configurable!)
    [ ] site configuration specifies pagination etc

[ ] if a file was updated, i.e. has a newer timestamp (and different checksum):
    rebuild it (but not more than once per minute)
[ ] if a file or directory was added then support the route
[ ] if a file or directory was removed then drop the route
[ ] make sure to also rebuild the generated pages etc
