# Play View Extractor

Extracts the largest (layer 0) image files from **PlayStation PlayView** files (`gvd.dat`). Exports as PNG to avoid compression.

# Usage

```
$ playview-extractor -h

  -debug
        output more log data
  -hidden
        whether to show the hidden areas (default true)
  -in string
        path to gvd.dat (default "gvd.dat")
  -layer int
        Target layer to export (default 0)
  -merge
        Whether to merge images to a combined image (default true)
  -out string
        output directory (default "out")
  -page string
        Target page to export (empty string exports all)

```

Put the executable into the folder with the `gvd.dat` file. All images are exported as `<filename>.png` without 
compression.

```
$ playview-extractor  
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
