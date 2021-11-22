# Halp
halp is a simple cli tool written in Go for small helper functions.

## Installing
#### Mac OSX
To install/update halp on an OSX build, you can do so using Homebrew. 
If you have not already, add the halp tap to your brew taps:
```$xslt
brew tap josh5276/halp git@github.com:josh5276/halp
```
Once tapped, you can install/upgrade/remove halp using the regular brew methods
```$xslt
brew update && brew install halp
```
```$xslt
brew update && brew upgrade halp
```
```$xslt
brew uninstall halp
```

## Contributing
#### Test this application 
* Run all tests
    * ```make tests```
* Run linting 
	* ```make lint```
    
#### Release a new build
The release for halp includes a few things:
1. Building of the Go binary 
2. Debian build for Linux installations
3. Brew build for OSX installations (this includes the formula.rb generation)
4. Archive build of tar.gz files
6. Drafting of a new release in github.

Once you have completed the prereqs, you should be able to draft a new release using make:
  * ```make release```  
    Note: this requires a clean git status and a new git tag.  If you are wanting to
    test this release, run ```make testrelease```       

#### Considerations
Halp uses a go-key library that has a dependency on an Go Sqlite library. 
Sqlite requires C and has a dependency on the GCC build, which is not compatible.

To get around this, you can manually install the linux gcc libraries by running

```brew install FiloSottile/musl-cross/musl-cross```
* NOTE: the goreleaser file already contains the arguments to reference the 
linux-gcc builds.

See: https://github.com/mattn/go-sqlite3/issues/384 for more information.

## More help
> Send an email to [Me!](mailto:josh.silvas@networktocode.com)
