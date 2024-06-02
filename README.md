# Play View Extractor

Extracts the largest (layer 0) image files from **PlayStation PlayView** files (`gvd.dat`). Only the jpegs will be 
exported. Other data is not supported.

# Usage

> For the current version, no configuration is available.

Put the executable into the folder with the `gvd.dat` file. All images are exported as `out_page_<n>.png` without 
compression.

```
$ play-view-extractor  
```

# Install 

You can use golang to build from source and install the extractor locally.

```
go install https://github.com/joernlenoch/playview-extractor/
```

# Build

This is a simple golang 1.22 project.

- Install GoLang 1.22
- Run `go build main.go`

# Special Thanks

Special thanks to the detailed file format information at https://www.psdevwiki.com/vita/PlayView
