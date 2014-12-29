// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package wicore implements all the interfaces to be shared between the wi
// editor process and its plugins.
//
// It is strongly versioned via the SHA-1 of its interface.
//
// Overview graph of the Editor and its hierarchical tree. In addition to
// knowing all Window as a tree, it has a direct link to every documents. This
// is because a document buffer may be loaded but not have any view associated,
// or two Views (panels) may be viewing the same Document.
//
//         +------+
//         |Editor|-------------+--------+
//         +------+             |        |
//             |                |        v
//             v                |   +--------+
//         +------+             |   |Document|
//         |Window|             |   +--------+
//         +------+             |
//            |                 |
//      +-----+-------+         |
//      |             |         |
//      v             v         |
//   +------+     +------+      |
//   |Window|     |Window|      |
//   +------+     +------+      |
//                   |          |
//                   v          |
//                 +----+       |
//                 |View|       |
//                 +----+       |
//                   |          |
//                   v          |
//               +--------+     |
//               |Document|<----+
//               +--------+
//
//
// A View references Commands specific to this View, keyboard mapping specific
// to this View and potentially a Document. Many View do not have a Document
// associated, like Status, the Command View, etc.
//
//              +----+
//              |View|
//              +----+
//                |
//       +--------+---+------------+
//       |            |            |
//       v            v            v
//   +--------+ +-----------+ +--------+
//   |Commands| |KeyBindings| |Document|
//   +--------+ +-----------+ +--------+
//       |
//       |   +-------+
//       +-->|Command|
//       |   +-------+
//       |
//       |   +-------+
//       +-->|Command|
//           +-------+
//
// Objects in wi main process:
//
//  - Editor
//  - Window
//
// Objects that can be in either in the wi main process or implemented (and
// brokered to the main process) in a plugin process:
//
//  - View
//  - Document
//  - KeyBindings
//  - Commands
//  - Command
package wicore
