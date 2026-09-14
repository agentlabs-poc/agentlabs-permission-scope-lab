// Package wiring is the composition root, and the only place that imports both
// domains.
//
// Auth-AL declares the two-question interface it needs; the application registry
// provides a type that satisfies it. Neither imports the other — Go interfaces
// are structural, so the two domains meet here and nowhere else.
//
// That is what makes the adaptation clean: put a different implementation behind
// the port and only this package changes.
package wiring
