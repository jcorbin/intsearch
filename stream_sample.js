'use strict';

module.exports = StreamSample;

function StreamSample(stream, n, seed, callback) {
    if (!(this instanceof StreamSample)) {
        return new StreamSample(stream, n, seed, callback);
    }

    var self = this;

    self.stream = stream;
    self.n = n;
    self.random = PRNG(seed);
    self.callback = callback;
    self.sample = [];
    self.filled = false;

    self.stream
        .on('data', onData)
        .on('end', onDone)
        .on('error', onError);

    function onError(err) {
        self.onError(err);
    }

    function onDone() {
        self.onDone();
    }

    function onData(item) {
        self.onData(item);
    }
}

StreamSample.prototype.onData = function onData(item) {
    var score = this.random();

    if (!this.filled) {
        this.sample.push(new SampleRecord(score, item));
        if (this.sample.length >= this.n) {
            this.filled = true;
            this.heapify();
        }
    } else if (score > this.sample[0].score) {
        this.sample[0].score = score;
        this.sample[0].item = item;
        this.siftdown(0);
    }
};

StreamSample.prototype.onError = function onError(err) {
    this.callback(err, this.sample);
};

StreamSample.prototype.onDone = function onDone() {
    this.callback(null, this.sample);
};

StreamSample.prototype.heapify = function heapify() {
    var i = Math.floor(this.sample.length / 2 - 1);
    for (; i >= 0; i--) {
        this.siftdown(i);
    }
};

StreamSample.prototype.siftdown = function siftdown(i) {
    while (true) {
        var par = this.sample[i];

        // left
        var j = 2 * i + 1;
        if (j >= this.sample.length) {
            return;
        }

        // maybe right
        var child = this.sample[j];
        if (++j >= this.sample.length ||
            this.sample[j].score >= child.score) {
            j--;
        } else {
            child = this.sample[j];
        }

        if (child.score >= par.score) {
            return;
        }

        this.sample[i] = child;
        this.sample[j] = par;

        i = j;
    }
};

function SampleRecord(score, item) {
    this.score = score;
    this.item = item;
}

function PRNG(seed) {
    // glibc
    var a = 1103515245;
    var c = 12345;
    var m = Math.pow(2, 31);

    // // VB 6
    // var a = 1140671485;
    // var c = 12820163;
    // var m = Math.pow(2, 24);

    var n = seed;
    return function random() {
        n = (a * n + c) % m;
        return n;
    };
}
