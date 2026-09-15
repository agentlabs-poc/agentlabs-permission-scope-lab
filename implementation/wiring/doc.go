// Package wiring is the composition root, and the only place that imports both
// domains. Open assembles the service; nothing else needs to know the order the
// stores open in or which port satisfies which interface.
//
// Auth-AL declares the two-question interface it needs; the application registry
// provides a type that satisfies it. Neither imports the other — Go interfaces
// are structural, so the two domains meet here and nowhere else.
//
// That is what makes the adaptation clean: put a different implementation behind
// the port and only this package changes.
//
// The rule used to be a claim in this comment while the steps were carried out
// twice — once in a test, once in a command — where they could drift apart
// without anything noticing. TestNeitherDomainImportsTheOther is the first thing
// that fails when the claim stops being true.
package wiring
