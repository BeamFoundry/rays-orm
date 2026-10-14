// # Initialization
//
// This package should be imported from a main file, alongside rays :
//
//	package main
//
//	import (
//	  "github.com/BeamFoundry/rays"
//	  _ "github.com/BeamFoundry/rays-orm"
//	)
//
// # Packages
//
// This project divides its code between the following packages :
//
//   - [github.com/BeamFoundry/rays-orm/pkg/core]: Contains all the repository interfaces and objects.
package rays_orm //nolint:revive // Ignoring package naming hint

import (
	_ "github.com/BeamFoundry/rays-orm/internal/autoconfig" // autoconfig
)
