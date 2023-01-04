# State base implementation of Tic Tac Toe
This implementation aims to demonstrate what are differences 
between two approaches, and whenever one is better than the other

- One thing that they share is that state is explicit in both approaches.
  - Evensourcing build state from changes, and changes are introduce by tranitions.
    - Explicit changes
  - State machine build state from transitions, and changes are implicit
    - Implicit changes.
      - Those changes can be calculated, by diffing previous state with current state.
      - Since states are unions, they may have different shape, and diffing is possible.
        - Change between GameProgress(moves[1,2]) and GameProgress(moves[1,2,3]) is move[3]
        - Which can be represented easly as GameProgressChanges(moves[1,2], moves[1,2,3])
      - Disadvantage is that in evensourcing, stat is refined and is parto of building block.
      - When in dynamicly computed changes, are not used, for something miningful, they may become less useful.

- Usage of mkunion simplfy implementation.
  - Exchustive checks are done by compiler.
  - No need to write switch statements.
  - No need to write visitor pattern thanks to MushMatchCommand function that was generated.


TODO

- [ ] create simple web app that will use this implementation
- [ ] create leaderboard with results
  - Leverage opus store
  - Leverage opus query