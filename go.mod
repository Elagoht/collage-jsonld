// A collage plugin that emits schema.org structured data into the document head.
module github.com/Elagoht/collage-jsonld

go 1.26

require github.com/Elagoht/collage v0.1.0

// Until collage has a released tag, this points the require at a checkout beside
// this one. Delete it once there is a version to fetch — it is the only line in
// this file that assumes anything about where the framework lives on disk.
replace github.com/Elagoht/collage => ../collage
