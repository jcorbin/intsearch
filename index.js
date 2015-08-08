'use strict';

var assert = require('assert');
var extend = require('xtend');

function WordProblem(word1, word2, word3) {
    var self = this;

    var lenDiff = word3.length - word2.length;
    assert(word1.length === word2.length);
    assert(lenDiff === 0 || lenDiff === 1);

    self.word1 = word1;
    self.word2 = word2;
    self.word3 = word3;
    self.lenDiff = lenDiff;

    self.initials = {};
    self.rounds = [];
    self.numLetters = 0;
    self.base = 0;
    self.letters = [];

    self.setup();
}

WordProblem.prototype.setup = function setup() {
    var self = this;

    self.initials[self.word1[0]] = true;
    self.initials[self.word2[0]] = true;
    self.initials[self.word3[0]] = true;

    var i;
    var round;
    var seen = {};
    var lastCarry = null;
    for (i = 1; i <= self.word1.length; i++) {
        var let1 = self.word1[self.word1.length - i];
        var let2 = self.word2[self.word2.length - i];
        var let3 = self.word3[self.word3.length - i];

        self.rounds.push(round = {
            expectedRem: let3,
            values: [let1, let2], // [= let3 [+ let1 let2 ?carryIn]]
            carryOut: 'rem' + i,
            letters: []
        });

        if (lastCarry) {
            round.values.unshift(lastCarry);
        }

        addLetter(let1);
        addLetter(let2);
        addLetter(let3);

        if (!round.letters.length) {
            round.letters = null;
        }

        lastCarry = round.carryOut;
    }

    if (self.lenDiff)  {
        var finalLet = self.word3[0];

        self.rounds.push(round = {
            expectedRem: finalLet,
            values: [lastCarry],
            carryOut: null,
            letters: []
        });

        if (!seen[finalLet]) {
            addLetter(finalLet);
        }
    }

    function addLetter(letter) {
        if (!seen[letter]) {
            seen[letter] = true;
            round.letters.push(letter);
            ++self.numLetters;
        }
    }

    self.base = baseFor(self.numLetters);
    self.letters = Object.keys(seen);
};

WordProblem.prototype.init = function init() {
    var self = this;

    var state = new WordProblemState();
    return state.init(self);
};

function WordProblemState(state) {
    var self = this;

    self.score = state ? state.score : 0;
    self.problem = state ? state.problem : null;
    self.valid = true;
    self.complete = false;
    self.digits = state ? state.digits : [];
    self.values = state ? state.values : {};
    self.roundIndex = state ? state.roundIndex : 0;
    self.round = state ? state.round : null;
}

WordProblemState.prototype.init = function init(prob) {
    var self = this;

    self.problem = prob;
    self.round = self.problem.rounds[self.roundIndex] || null;

    for (var i = 0; i < self.problem.base; i++) {
        self.digits.push(i);
    }

    for (i = 0; i < self.problem.letters.length; i++) {
        var letter = self.problem.letters[i];
        self.values[letter] = null;
    }

    for (i = 0; i < self.problem.rounds.length; i++) {
        var round = self.problem.rounds[i];
        if (round.carryOut) {
            self.values[round.carryOut] = null;
        }
    }

    return self;
};

WordProblemState.prototype.successors = function successors() {
    var self = this;

    while (self.round && self.roundIndex < self.problem.rounds.length) {
        if (self.round.letters) {
            var succ = self.chooseNextRoundLetter();
            if (succ !== null) {
                return succ;
            }
        } else if (!self.checkConstraint()) {
            return [];
        }

        self.round = self.problem.rounds[++self.roundIndex] || null;
        console.log('advanced round', self.round);
    }

    self.complete = true;
    return [];
};

WordProblemState.prototype.chooseNextRoundLetter = function chooseNextRoundLetter() {
    var self = this;

    for (var i = 0; i < self.round.letters.length; i++) {
        var letter = self.round.letters[i];
        if (self.values[letter] !== null) {
            continue;
        }

        var succ = [];
        if (i === self.round.letters.length - 1) {
            self.chooseLetter(letter, function checkEach(next) {
                if (next.checkConstraint()) {
                    succ.push(next);
                }
            });
        } else {
            self.chooseLetter(letter, function takeEach(next) {
                succ.push(next);
            });
        }
        if (!succ.length) {
            self.valid = false;
        }
        return succ;
    }

    return null;
};

WordProblemState.prototype.chooseLetter = function chooseLetter(letter, each) {
    var self = this;

    var start = self.problem.initials[letter] ? 1 : 0;
    for (var i = start; i < self.digits.length; i++) {
        // TODO: borrow prior unvalid state from last round
        var next = new WordProblemState(self);
        next.values = extend(next.values);

        next.score++;
        next.values[letter] = self.digits[i];
        next.digits = self.digits.slice(0, i).concat(self.digits.slice(i + 1));
        each(next);
    }
};

WordProblemState.prototype.result = function result() {
    var self = this;

    var result = {};
    [
        self.problem.word1,
        self.problem.word2,
        self.problem.word3
    ].forEach(function toNumber(word) {
        var value = 0;
        for (var i = word.length-1; i >= 0; i--) {
            value = 10 * value + self.values[word[i]];
        }
        result[word] = value;
    });
    return result;
};

WordProblemState.prototype.valString = function valString() {
    var self = this;

    return self.problem.letters
        .filter(function nonNullVal(letter) {
            return self.values[letter] !== null;
        })
        .map(function letVal(letter) {
            return letter + ':' + self.values[letter];
        })
        .join(' ');
};

WordProblemState.prototype.checkConstraint = function checkConstraint() {
    var self = this;

    var expected = self.values[self.round.expectedRem];
    var value = 0;

    for (var i = 0; i < self.round.values.length; i++) {
        value += self.values[self.round.values[i]];
    }

    var letVals = self.round.values.map(function letVal(letter) {
        return self.values[letter];
    });

    if (self.round.carryOut) {
        var quo = Math.floor(value / self.problem.base);
        var rem = value % self.problem.base;

        self.values[self.round.carryOut] = quo;
        self.valid = rem === expected;

        if (self.valid) {
            console.log('PASSED', self.round, self.valString(), letVals);
        }

    } else {
        self.valid = value === expected;

        console.log('LAST', self.valid, self.round, self.valString(), letVals);
    }

    return self.valid;
};

/*
 *     S E N D
 * +   M O R E
 * -----------
 *   M O N E Y
 */

var util = require('util');

function main() {
    var prob = new WordProblem('send', 'more', 'money');

    search(prob.init(), function each(res) {
        var valid = res.send + res.more === res.money;
        console.log(valid ? 'GOOD' : 'BAD', res);

        return valid;
    }, function trace(state, succ) {
        if (!succ.length || !state.valid) return;
        console.log('expanded %s from score:%s complete:%s valid:%s values: %s',
            succ.length,
            state.score, state.complete, state.valid,
            state.valString()
        );
    });
}

function search(state, each, trace) {
    if (!trace) {
        trace = noop;
    }

    var frontier = [state];
    while (frontier.length) {
        state = frontier.shift();

        var succ = state.successors();
        if (frontier.length) {
            frontier.push.apply(frontier, succ);
        } else {
            frontier = succ;
        }

        trace(state, succ);

        heapify(frontier);

        if (state.complete && state.valid) {
            var res = state.result();
            if (each(res)) {
                break;
            }
        }
    }
}

function heapify(array) {
    for (var i = Math.floor(array.length / 2 - 1); i >= 0; i--) {
        siftdown(array, i);
    }
}

function siftdown(array, i) {
    while (true) {
        var left = (2 * i) + 1;
        if (left >= array.length) {
            return;
        }

        var right = left + 1;
        var child = left;
        if (right < array.length &&
            array[right].score > array[left].score) {
            child = right;
        }

        if (array[child].score <= array[i].score) {
            return;
        }

        var tmp = array[child];
        array[child] = array[i];
        array[i] = tmp;
        i = child;
    }
}

main();

function baseFor(n) {
    if (n < 10) {
        return 10;
    } else if (n < 16) {
        return 16;
    } else if (n < 32) {
        return 32;
    } else if (n < 64) {
        return 64;
    } else {
        assert(false, 'not supported');
    }
}

function noop() {
}
