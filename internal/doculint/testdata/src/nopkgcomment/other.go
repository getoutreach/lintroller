// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: A package whose comment lives in no file named after the package.

package nopkgcomment // want `has no file with the same name containing package comment`

// Thing is a thing.
func Thing() {}
