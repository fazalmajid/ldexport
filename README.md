# ldexport
Export secrets from the Lockdown Mac/iOS two-factor authentication app

I use the [Lockdown](http://cocoaapp.com/lockdown/) app to manage my TOTP two-factor authentication on my iPhone, iPad and Mac. While it has a convenient iCloud-based sync functionality to keep all of these devices in sync, it does not have a convenient way to export all the secrets for backup or migration purposes.

ldexport is a tool to fix that. It runs on the Mac and exports the app's secrets in either:

* JSON
* HTML (suitable for printing as hardcopy backup, e.g. as part of an emergency [in-case of](https://web.archive.org/web/20150411123043/http://unclutterer.com/2011/08/16/creating-an-in-case-of-file/) file) (although you may want to take precautions in case you are burglarized, e.g. put it in a safe deposit box or leave it with your attorney).

## Building

You need Go installed (tested with Go 1.12.7 and 1.15.5):

```go get github.com/fazalmajid/ldexport```

your `$GOPATH/bin` will have a single executable `ldexport`

I have included a binary version (compiled using Go 1.15.5 on 10.14.6 Mojave) for those who don't have Go, but it's not good or safe practice to rely on binary software from some random person on the Internet for such security-critical data...

The checksums are:

```
fafnir ~/ldexport>gsha1sum ldexport
6ea3a5931cc74ed71a4413da84f92437d9e20154  ldexport
fafnir ~/ldexport>gsha256sum ldexport
1fab84e681886abdda51cd41d24a8e325b07a24399822ceec96d6b97069efc96  ldexport
```

## Usage

```
Usage of ./ldexport:
  -a	also include archived secrets
  -html
    	export in HTML format
```
By default it exports to JSON, but with `-html` it will output a self-contained single-file HTML to standard output.

It will not export Lockdown's "archived" secrets by default, because presumably you archived them for a reason, but adding the `-a` flag will also include them.

## Sample JSON output

(fake credentials, of course)

```yaml
[
  {
    "Service": "Amazon",
    "Login": "amazon@example.com",
    "Created": "2015-11-18T19:53:34.969532012Z",
    "Modified": "2015-11-18T19:53:34.969532012Z",
    "URL": "otpauth://totp/Amazon%3Aamazon%40example.com?secret=M7IoBWqA2WuzYG27ju82XTWsflPEha3xBafMQ3i9CgwKgp6RdBGh&issuer=Amazon",
    "Favorite": true,
    "Archived": false
  },
  {
    "Service": "PayPal",
    "Login": "ebay@example.com",
    "Created": "2019-11-25T08:46:57.253684043Z",
    "Modified": "2019-11-25T08:46:57.253684043Z",
    "URL": "otpauth://totp/PayPal:ebay@example.com?secret=3gB0VWJFkaYcVIiD&issuer=PayPal",
    "Favorite": false,
    "Archived": false
  },
  {
    "Service": "Reddit",
    "Login": "johndoe",
    "Created": "2020-08-07T19:58:37.930042982+01:00",
    "Modified": "2020-08-07T19:58:37.930042982+01:00",
    "URL": "otpauth://totp/Reddit:johndoe?secret=nDTxDMI6bEgVpHWCViZjDFhXKH1bysRa&issuer=Reddit",
    "Favorite": true,
    "Archived": false
  },
  {
    "Service": "GitHub",
    "Login": "",
    "Created": "2016-05-04T19:04:12.495128989+01:00",
    "Modified": "2017-04-04T06:33:10.641680002+01:00",
    "URL": "otpauth://totp/github.com/johndoe?issuer=GitHub&secret=bXh5qmeTMzcatKKz",
    "Favorite": false,
    "Archived": false
  },
  {
    "Service": "Google",
    "Login": "johndoe@gmail.com",
    "Created": "2015-11-13T05:06:07.103500008Z",
    "Modified": "2015-11-13T05:06:07.103500008Z",
    "URL": "otpauth://totp/Google%3Ajohndoe%40gmail.com?secret=o5MvqdWDt7ZEHHSTuH6rCAUr4M6ozGQD&issuer=Google",
    "Favorite": false,
    "Archived": false
  }
]
```

## Technical details

The Mac version of Lockdown saves its secrets in a plist file, which in turn contains a nested plist file in Apple's crackpot `NSKeyedArchiver` format. Fortunately 

## Credits

* [Sarah Edwards](https://www.linkedin.com/in/sledwards/) for [documenting](https://www.mac4n6.com/blog/2016/1/1/manual-analysis-of-nskeyedarchiver-formatted-plist-files-a-review-of-the-new-os-x-1011-recent-items) Apple's crackpot NSKeyedArchiver format
* [Dustin L. Howett](https://github.com/DHowett) for his [Go plist module](https://github.com/DHowett/go-plist/)
* [Russ Cox](https://swtch.com/~rsc/) for his [Go QR code module](https://godoc.org/rsc.io/qr)
