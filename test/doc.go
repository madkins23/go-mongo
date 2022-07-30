// Package test provides resources to tests in other packages.
// Can't name files as *_test.go as they won't link properly from this package.
//
// Hopefully this package won't be linked into the eventual module/application.
// Some experimentation using code from:
//  https://stackoverflow.com/questions/70764915/how-to-check-the-size-of-packages-linked-into-my-go-code
// seems to demonstrate that this is true.
// Build the code from that link and run it against an application that uses
// a package (other than this one) from this project to test it yourself.
package test
