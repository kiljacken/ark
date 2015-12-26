On modules strings (a::b::c) for build paths:
    The user really shouldn't have to do this conversion himself. We could
    trivially compare the passed path to the base path and build a module
        name from that.

    Another thing to consider is how we're going to handle passing multiple
    files when building a binary. In which module do we look for main?
    There's probably also a load of other considerations that relate to
    this problem.

On directory modules for building:
	To handle modules being a complete directory istead of a single file we
	need first of all need to figure out how to represent this internally.
	Having though about the problems the following comes to mind, in order
	of when in the compilation process the change would take place:

		1. Simply merge the parse trees within a given module.
			- This should result in a fairly simple way of handling the
			parsing aspect, but doesn't really feel robust to me.

		2. Make the constructor take multiple parse trees.
			- The constructor would then create an AST based on all the
			parse trees. A lot better than 1, not sure if there's any
			downsides to this approach.

		3. Make the `Module` struct store several ASTs.
			- This would require major changes to the way scopes are
			handled, as several ASTs would share a scope etc. This would be
			the biggest overhaul between the options, and I'm not sure if
			it brings much to the table in terms of safety above the other
			methods.

	Having written this out it seems like #2 will probably be the best
	solution, both in terms of simplicity and robustness.


	Another thing that should be considered is the level of scope
	seperation. If I were to issue an `use abc` in a file that is part of
	module `a`, and then in another file from the same module do the same,
	should that produce an error? Also the other side of the case: if we
	use `abc` in one file contained by module `a`, should I be able to
	access it from other files in abc?

	The answer to this is probably that we should scope `use`s on a per
	file basis, such that we don't pollute the namespace of the module.