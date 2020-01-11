// Package kickoff provides a CLI tool for bootstrapping projects from skeleton
// directories. The skeletons can contain any kind of files and special *.skel
// files which will be evaluated using the golang template engine.
//
// The tool was initial meant for bootstrapping golang projects, but it is
// actually language agnostic as the skeletons do not need to be golang
// specific.
package kickoff
