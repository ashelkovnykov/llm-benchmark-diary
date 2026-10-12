## Summary

**Time:** ~10 hours \
**Cost:** free \
**App Quality:** 6/10 \
**Code Quality:** 14/20

A hand-written solution for the experiment task to serve as a benchmark. More than half of the time spent writing it was spent reading the Bubble Tea documentation. Knowing what I now know about Bubble Tea, I probably wouldn't write it this way again, referring to the nested `tea.Model` interface pattern I used. The fact that `Model.Update()` returns a `tea.Model` made me assume that passing structs implementing that interface back and forth was the intended design choice, but I was apparently mistaken and several issues made this a rather annoying process:
1. Bubble Tea doesn't provide any direct method for defining parent or sibling relationships between `tea.Model`s. It also doesn't have any way of registering multiple `tea.Model`s with the `tea.Program`. This means boilerplate in each struct that extends `tea.Model` to manually manage these relationships.
2. Some pre-built bubbles _require_ `Model.Init()` to be called in order to function correctly. However, this function is only automatically called _exactly once_, and only on the specific `tea.Model` registered with `tea.Program`. This means manually calling `Model.Init()` for those bubbles and properly handling the output.
3. Some pre-built bubbles _require_ the terminal size to be known at initialization time, or to be provided shortly thereafter by a `tea.WindowSizeMessage`. `tea.Program` will actually automatically send such a message on start-up - but, again: _exaclty once_ and _only_ to the `tea.Model` registered with `tea.Program`. This means more boilerplate to properly forward these messages.
4. The order of these messages arriving to your `tea.Model` is _non-deterministic_. In rare cases, I would even see keyboard input arriving before `tea.WindowSizeMessage` or messages stemming from `Model.Init()`.

It appears that the idiomatic solution is to instead use a parent structure that both implements `tea.Model` and _contains_ child structs extending `tea.Model`, then route between them using an internal state machine.
