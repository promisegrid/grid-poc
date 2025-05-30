"use strict";
(function() {

var $goVersion = "go1.19.13";
Error.stackTraceLimit = Infinity;

var $NaN = NaN;
var $global, $module;
if (typeof window !== "undefined") { /* web page */
    $global = window;
} else if (typeof self !== "undefined") { /* web worker */
    $global = self;
} else if (typeof global !== "undefined") { /* Node.js */
    $global = global;
    $global.require = require;
} else { /* others (e.g. Nashorn) */
    $global = this;
}

if ($global === undefined || $global.Array === undefined) {
    throw new Error("no global object found");
}
if (typeof module !== "undefined") {
    $module = module;
}

if (!$global.fs && $global.require) {
    try {
        var fs = $global.require('fs');
        if (typeof fs === "object" && fs !== null && Object.keys(fs).length !== 0) {
            $global.fs = fs;
        }
    } catch (e) { /* Ignore if the module couldn't be loaded. */ }
}

if (!$global.fs) {
    var outputBuf = "";
    var decoder = new TextDecoder("utf-8");
    $global.fs = {
        constants: { O_WRONLY: -1, O_RDWR: -1, O_CREAT: -1, O_TRUNC: -1, O_APPEND: -1, O_EXCL: -1 }, // unused
        writeSync: function writeSync(fd, buf) {
            outputBuf += decoder.decode(buf);
            var nl = outputBuf.lastIndexOf("\n");
            if (nl != -1) {
                console.log(outputBuf.substr(0, nl));
                outputBuf = outputBuf.substr(nl + 1);
            }
            return buf.length;
        },
        write: function write(fd, buf, offset, length, position, callback) {
            if (offset !== 0 || length !== buf.length || position !== null) {
                callback(enosys());
                return;
            }
            var n = this.writeSync(fd, buf);
            callback(null, n);
        }
    };
}

var $linknames = {} // Collection of functions referenced by a go:linkname directive.
var $packages = {}, $idCounter = 0;
var $keys = m => { return m ? Object.keys(m) : []; };
var $flushConsole = () => { };
var $throwRuntimeError; /* set by package "runtime" */
var $throwNilPointerError = () => { $throwRuntimeError("invalid memory address or nil pointer dereference"); };
var $call = (fn, rcvr, args) => { return fn.apply(rcvr, args); };
var $makeFunc = fn => { return function(...args) { return $externalize(fn(this, new ($sliceType($jsObjectPtr))($global.Array.prototype.slice.call(args, []))), $emptyInterface); }; };
var $unused = v => { };
var $print = console.log;
// Under Node we can emulate print() more closely by avoiding a newline.
if (($global.process !== undefined) && $global.require) {
    try {
        var util = $global.require('util');
        $print = function(...args) { $global.process.stderr.write(util.format.apply(this, args)); };
    } catch (e) {
        // Failed to require util module, keep using console.log().
    }
}
var $println = console.log

var $initAllLinknames = () => {
    var names = $keys($packages);
    for (var i = 0; i < names.length; i++) {
        var f = $packages[names[i]]["$initLinknames"];
        if (typeof f == 'function') {
            f();
        }
    }
}

var $mapArray = (array, f) => {
    var newArray = new array.constructor(array.length);
    for (var i = 0; i < array.length; i++) {
        newArray[i] = f(array[i]);
    }
    return newArray;
};

// $mapIndex returns the value of the given key in m, or undefined if m is nil/undefined or not a map
var $mapIndex = (m, key) => {
    return typeof m.get === "function" ? m.get(key) : undefined;
};
// $mapDelete deletes the key and associated value from m.  If m is nil/undefined or not a map, $mapDelete is a no-op
var $mapDelete = (m, key) => {
    typeof m.delete === "function" && m.delete(key)
};
// Returns a method bound to the receiver instance, safe to invoke as a 
// standalone function. Bound function is cached for later reuse.
var $methodVal = (recv, name) => {
    var vals = recv.$methodVals || {};
    recv.$methodVals = vals; /* noop for primitives */
    var f = vals[name];
    if (f !== undefined) {
        return f;
    }
    var method = recv[name];
    f = method.bind(recv);
    vals[name] = f;
    return f;
};

var $methodExpr = (typ, name) => {
    var method = typ.prototype[name];
    if (method.$expr === undefined) {
        method.$expr = (...args) => {
            $stackDepthOffset--;
            try {
                if (typ.wrapped) {
                    args[0] = new typ(args[0]);
                }
                return Function.call.apply(method, args);
            } finally {
                $stackDepthOffset++;
            }
        };
    }
    return method.$expr;
};

var $ifaceMethodExprs = {};
var $ifaceMethodExpr = name => {
    var expr = $ifaceMethodExprs["$" + name];
    if (expr === undefined) {
        expr = $ifaceMethodExprs["$" + name] = (...args) => {
            $stackDepthOffset--;
            try {
                return Function.call.apply(args[0][name], args);
            } finally {
                $stackDepthOffset++;
            }
        };
    }
    return expr;
};

var $subslice = (slice, low, high, max) => {
    if (high === undefined) {
        high = slice.$length;
    }
    if (max === undefined) {
        max = slice.$capacity;
    }
    if (low < 0 || high < low || max < high || high > slice.$capacity || max > slice.$capacity) {
        $throwRuntimeError("slice bounds out of range");
    }
    if (slice === slice.constructor.nil) {
        return slice;
    }
    var s = new slice.constructor(slice.$array);
    s.$offset = slice.$offset + low;
    s.$length = high - low;
    s.$capacity = max - low;
    return s;
};

var $substring = (str, low, high) => {
    if (low < 0 || high < low || high > str.length) {
        $throwRuntimeError("slice bounds out of range");
    }
    return str.substring(low, high);
};

// Convert Go slice to an equivalent JS array type.
var $sliceToNativeArray = slice => {
    if (slice.$array.constructor !== Array) {
        return slice.$array.subarray(slice.$offset, slice.$offset + slice.$length);
    }
    return slice.$array.slice(slice.$offset, slice.$offset + slice.$length);
};

// Convert Go slice to a pointer to an underlying Go array.
// 
// Note that an array pointer can be represented by an "unwrapped" native array
// type, and it will be wrapped back into its Go type when necessary.
var $sliceToGoArray = (slice, arrayPtrType) => {
    var arrayType = arrayPtrType.elem;
    if (arrayType !== undefined && slice.$length < arrayType.len) {
        $throwRuntimeError("cannot convert slice with length " + slice.$length + " to pointer to array with length " + arrayType.len);
    }
    if (slice == slice.constructor.nil) {
        return arrayPtrType.nil; // Nil slice converts to nil array pointer.
    }
    if (slice.$array.constructor !== Array) {
        return slice.$array.subarray(slice.$offset, slice.$offset + arrayType.len);
    }
    if (slice.$offset == 0 && slice.$length == slice.$capacity && slice.$length == arrayType.len) {
        return slice.$array;
    }
    if (arrayType.len == 0) {
        return new arrayType([]);
    }

    // Array.slice (unlike TypedArray.subarray) returns a copy of an array range,
    // which is not sharing memory with the original one, which violates the spec
    // for slice to array conversion. This is incompatible with the Go spec, in
    // particular that the assignments to the array elements would be visible in
    // the slice. Prefer to fail explicitly instead of creating subtle bugs.
    $throwRuntimeError("gopherjs: non-numeric slice to underlying array conversion is not supported for subslices");
};

// Convert between compatible slice types (e.g. native and names).
var $convertSliceType = (slice, desiredType) => {
    if (slice == slice.constructor.nil) {
        return desiredType.nil; // Preserve nil value.
    }

    return $subslice(new desiredType(slice.$array), slice.$offset, slice.$offset + slice.$length);
}

var $decodeRune = (str, pos) => {
    var c0 = str.charCodeAt(pos);

    if (c0 < 0x80) {
        return [c0, 1];
    }

    if (c0 !== c0 || c0 < 0xC0) {
        return [0xFFFD, 1];
    }

    var c1 = str.charCodeAt(pos + 1);
    if (c1 !== c1 || c1 < 0x80 || 0xC0 <= c1) {
        return [0xFFFD, 1];
    }

    if (c0 < 0xE0) {
        var r = (c0 & 0x1F) << 6 | (c1 & 0x3F);
        if (r <= 0x7F) {
            return [0xFFFD, 1];
        }
        return [r, 2];
    }

    var c2 = str.charCodeAt(pos + 2);
    if (c2 !== c2 || c2 < 0x80 || 0xC0 <= c2) {
        return [0xFFFD, 1];
    }

    if (c0 < 0xF0) {
        var r = (c0 & 0x0F) << 12 | (c1 & 0x3F) << 6 | (c2 & 0x3F);
        if (r <= 0x7FF) {
            return [0xFFFD, 1];
        }
        if (0xD800 <= r && r <= 0xDFFF) {
            return [0xFFFD, 1];
        }
        return [r, 3];
    }

    var c3 = str.charCodeAt(pos + 3);
    if (c3 !== c3 || c3 < 0x80 || 0xC0 <= c3) {
        return [0xFFFD, 1];
    }

    if (c0 < 0xF8) {
        var r = (c0 & 0x07) << 18 | (c1 & 0x3F) << 12 | (c2 & 0x3F) << 6 | (c3 & 0x3F);
        if (r <= 0xFFFF || 0x10FFFF < r) {
            return [0xFFFD, 1];
        }
        return [r, 4];
    }

    return [0xFFFD, 1];
};

var $encodeRune = r => {
    if (r < 0 || r > 0x10FFFF || (0xD800 <= r && r <= 0xDFFF)) {
        r = 0xFFFD;
    }
    if (r <= 0x7F) {
        return String.fromCharCode(r);
    }
    if (r <= 0x7FF) {
        return String.fromCharCode(0xC0 | r >> 6, 0x80 | (r & 0x3F));
    }
    if (r <= 0xFFFF) {
        return String.fromCharCode(0xE0 | r >> 12, 0x80 | (r >> 6 & 0x3F), 0x80 | (r & 0x3F));
    }
    return String.fromCharCode(0xF0 | r >> 18, 0x80 | (r >> 12 & 0x3F), 0x80 | (r >> 6 & 0x3F), 0x80 | (r & 0x3F));
};

var $stringToBytes = str => {
    var array = new Uint8Array(str.length);
    for (var i = 0; i < str.length; i++) {
        array[i] = str.charCodeAt(i);
    }
    return array;
};

var $bytesToString = slice => {
    if (slice.$length === 0) {
        return "";
    }
    var str = "";
    for (var i = 0; i < slice.$length; i += 10000) {
        str += String.fromCharCode.apply(undefined, slice.$array.subarray(slice.$offset + i, slice.$offset + Math.min(slice.$length, i + 10000)));
    }
    return str;
};

var $stringToRunes = str => {
    var array = new Int32Array(str.length);
    var rune, j = 0;
    for (var i = 0; i < str.length; i += rune[1], j++) {
        rune = $decodeRune(str, i);
        array[j] = rune[0];
    }
    return array.subarray(0, j);
};

var $runesToString = slice => {
    if (slice.$length === 0) {
        return "";
    }
    var str = "";
    for (var i = 0; i < slice.$length; i++) {
        str += $encodeRune(slice.$array[slice.$offset + i]);
    }
    return str;
};

var $copyString = (dst, src) => {
    var n = Math.min(src.length, dst.$length);
    for (var i = 0; i < n; i++) {
        dst.$array[dst.$offset + i] = src.charCodeAt(i);
    }
    return n;
};

var $copySlice = (dst, src) => {
    var n = Math.min(src.$length, dst.$length);
    $copyArray(dst.$array, src.$array, dst.$offset, src.$offset, n, dst.constructor.elem);
    return n;
};

var $copyArray = (dst, src, dstOffset, srcOffset, n, elem) => {
    if (n === 0 || (dst === src && dstOffset === srcOffset)) {
        return;
    }

    if (src.subarray) {
        dst.set(src.subarray(srcOffset, srcOffset + n), dstOffset);
        return;
    }

    switch (elem.kind) {
        case $kindArray:
        case $kindStruct:
            if (dst === src && dstOffset > srcOffset) {
                for (var i = n - 1; i >= 0; i--) {
                    elem.copy(dst[dstOffset + i], src[srcOffset + i]);
                }
                return;
            }
            for (var i = 0; i < n; i++) {
                elem.copy(dst[dstOffset + i], src[srcOffset + i]);
            }
            return;
    }

    if (dst === src && dstOffset > srcOffset) {
        for (var i = n - 1; i >= 0; i--) {
            dst[dstOffset + i] = src[srcOffset + i];
        }
        return;
    }
    for (var i = 0; i < n; i++) {
        dst[dstOffset + i] = src[srcOffset + i];
    }
};

var $clone = (src, type) => {
    var clone = type.zero();
    type.copy(clone, src);
    return clone;
};

var $pointerOfStructConversion = (obj, type) => {
    if (obj.$proxies === undefined) {
        obj.$proxies = {};
        obj.$proxies[obj.constructor.string] = obj;
    }
    var proxy = obj.$proxies[type.string];
    if (proxy === undefined) {
        var properties = {};
        for (var i = 0; i < type.elem.fields.length; i++) {
            (fieldProp => {
                properties[fieldProp] = {
                    get() { return obj[fieldProp]; },
                    set(value) { obj[fieldProp] = value; }
                };
            })(type.elem.fields[i].prop);
        }
        proxy = Object.create(type.prototype, properties);
        proxy.$val = proxy;
        obj.$proxies[type.string] = proxy;
        proxy.$proxies = obj.$proxies;
    }
    return proxy;
};

var $append = function (slice) {
    return $internalAppend(slice, arguments, 1, arguments.length - 1);
};

var $appendSlice = (slice, toAppend) => {
    if (toAppend.constructor === String) {
        var bytes = $stringToBytes(toAppend);
        return $internalAppend(slice, bytes, 0, bytes.length);
    }
    return $internalAppend(slice, toAppend.$array, toAppend.$offset, toAppend.$length);
};

var $internalAppend = (slice, array, offset, length) => {
    if (length === 0) {
        return slice;
    }

    var newArray = slice.$array;
    var newOffset = slice.$offset;
    var newLength = slice.$length + length;
    var newCapacity = slice.$capacity;

    if (newLength > newCapacity) {
        newOffset = 0;
        newCapacity = Math.max(newLength, slice.$capacity < 1024 ? slice.$capacity * 2 : Math.floor(slice.$capacity * 5 / 4));

        if (slice.$array.constructor === Array) {
            newArray = slice.$array.slice(slice.$offset, slice.$offset + slice.$length);
            newArray.length = newCapacity;
            var zero = slice.constructor.elem.zero;
            for (var i = slice.$length; i < newCapacity; i++) {
                newArray[i] = zero();
            }
        } else {
            newArray = new slice.$array.constructor(newCapacity);
            newArray.set(slice.$array.subarray(slice.$offset, slice.$offset + slice.$length));
        }
    }

    $copyArray(newArray, array, newOffset + slice.$length, offset, length, slice.constructor.elem);

    var newSlice = new slice.constructor(newArray);
    newSlice.$offset = newOffset;
    newSlice.$length = newLength;
    newSlice.$capacity = newCapacity;
    return newSlice;
};

var $equal = (a, b, type) => {
    if (type === $jsObjectPtr) {
        return a === b;
    }
    switch (type.kind) {
        case $kindComplex64:
        case $kindComplex128:
            return a.$real === b.$real && a.$imag === b.$imag;
        case $kindInt64:
        case $kindUint64:
            return a.$high === b.$high && a.$low === b.$low;
        case $kindArray:
            if (a.length !== b.length) {
                return false;
            }
            for (var i = 0; i < a.length; i++) {
                if (!$equal(a[i], b[i], type.elem)) {
                    return false;
                }
            }
            return true;
        case $kindStruct:
            for (var i = 0; i < type.fields.length; i++) {
                var f = type.fields[i];
                if (!$equal(a[f.prop], b[f.prop], f.typ)) {
                    return false;
                }
            }
            return true;
        case $kindInterface:
            return $interfaceIsEqual(a, b);
        default:
            return a === b;
    }
};

var $interfaceIsEqual = (a, b) => {
    if (a === $ifaceNil || b === $ifaceNil) {
        return a === b;
    }
    if (a.constructor !== b.constructor) {
        return false;
    }
    if (a.constructor === $jsObjectPtr) {
        return a.object === b.object;
    }
    if (!a.constructor.comparable) {
        $throwRuntimeError("comparing uncomparable type " + a.constructor.string);
    }
    return $equal(a.$val, b.$val, a.constructor);
};

var $unsafeMethodToFunction = (typ, name, isPtr) => {
    if (isPtr) {
        return (r, ...args) => {
            var ptrType = $ptrType(typ);
            if (r.constructor != ptrType) {
                switch (typ.kind) {
                    case $kindStruct:
                        r = $pointerOfStructConversion(r, ptrType);
                        break;
                    case $kindArray:
                        r = new ptrType(r);
                        break;
                    default:
                        r = new ptrType(r.$get, r.$set, r.$target);
                }
            }
            return r[name](...args);
        };
    } else {
        return (r, ...args) => {
            var ptrType = $ptrType(typ);
            if (r.constructor != ptrType) {
                switch (typ.kind) {
                    case $kindStruct:
                        r = $clone(r, typ);
                        break;
                    case $kindSlice:
                        r = $convertSliceType(r, typ);
                        break;
                    case $kindComplex64:
                    case $kindComplex128:
                        r = new typ(r.$real, r.$imag);
                        break;
                    default:
                        r = new typ(r);
                }
            }
            return r[name](...args);
        };
    }
};

var $id = x => {
    return x;
};

var $instanceOf = (x, y) => {
    return x instanceof y;
};

var $typeOf = x => {
    return typeof (x);
};
var $min = Math.min;
var $mod = (x, y) => { return x % y; };
var $parseInt = parseInt;
var $parseFloat = f => {
    if (f !== undefined && f !== null && f.constructor === Number) {
        return f;
    }
    return parseFloat(f);
};

var $froundBuf = new Float32Array(1);
var $fround = Math.fround || (f => {
    $froundBuf[0] = f;
    return $froundBuf[0];
});

var $imul = Math.imul || ((a, b) => {
    var ah = (a >>> 16) & 0xffff;
    var al = a & 0xffff;
    var bh = (b >>> 16) & 0xffff;
    var bl = b & 0xffff;
    return ((al * bl) + (((ah * bl + al * bh) << 16) >>> 0) >> 0);
});

var $floatKey = f => {
    if (f !== f) {
        $idCounter++;
        return "NaN$" + $idCounter;
    }
    return String(f);
};

var $flatten64 = x => {
    return x.$high * 4294967296 + x.$low;
};

var $shiftLeft64 = (x, y) => {
    if (y === 0) {
        return x;
    }
    if (y < 32) {
        return new x.constructor(x.$high << y | x.$low >>> (32 - y), (x.$low << y) >>> 0);
    }
    if (y < 64) {
        return new x.constructor(x.$low << (y - 32), 0);
    }
    return new x.constructor(0, 0);
};

var $shiftRightInt64 = (x, y) => {
    if (y === 0) {
        return x;
    }
    if (y < 32) {
        return new x.constructor(x.$high >> y, (x.$low >>> y | x.$high << (32 - y)) >>> 0);
    }
    if (y < 64) {
        return new x.constructor(x.$high >> 31, (x.$high >> (y - 32)) >>> 0);
    }
    if (x.$high < 0) {
        return new x.constructor(-1, 4294967295);
    }
    return new x.constructor(0, 0);
};

var $shiftRightUint64 = (x, y) => {
    if (y === 0) {
        return x;
    }
    if (y < 32) {
        return new x.constructor(x.$high >>> y, (x.$low >>> y | x.$high << (32 - y)) >>> 0);
    }
    if (y < 64) {
        return new x.constructor(0, x.$high >>> (y - 32));
    }
    return new x.constructor(0, 0);
};

var $mul64 = (x, y) => {
    var x48 = x.$high >>> 16;
    var x32 = x.$high & 0xFFFF;
    var x16 = x.$low >>> 16;
    var x00 = x.$low & 0xFFFF;

    var y48 = y.$high >>> 16;
    var y32 = y.$high & 0xFFFF;
    var y16 = y.$low >>> 16;
    var y00 = y.$low & 0xFFFF;

    var z48 = 0, z32 = 0, z16 = 0, z00 = 0;
    z00 += x00 * y00;
    z16 += z00 >>> 16;
    z00 &= 0xFFFF;
    z16 += x16 * y00;
    z32 += z16 >>> 16;
    z16 &= 0xFFFF;
    z16 += x00 * y16;
    z32 += z16 >>> 16;
    z16 &= 0xFFFF;
    z32 += x32 * y00;
    z48 += z32 >>> 16;
    z32 &= 0xFFFF;
    z32 += x16 * y16;
    z48 += z32 >>> 16;
    z32 &= 0xFFFF;
    z32 += x00 * y32;
    z48 += z32 >>> 16;
    z32 &= 0xFFFF;
    z48 += x48 * y00 + x32 * y16 + x16 * y32 + x00 * y48;
    z48 &= 0xFFFF;

    var hi = ((z48 << 16) | z32) >>> 0;
    var lo = ((z16 << 16) | z00) >>> 0;

    var r = new x.constructor(hi, lo);
    return r;
};

var $div64 = (x, y, returnRemainder) => {
    if (y.$high === 0 && y.$low === 0) {
        $throwRuntimeError("integer divide by zero");
    }

    var s = 1;
    var rs = 1;

    var xHigh = x.$high;
    var xLow = x.$low;
    if (xHigh < 0) {
        s = -1;
        rs = -1;
        xHigh = -xHigh;
        if (xLow !== 0) {
            xHigh--;
            xLow = 4294967296 - xLow;
        }
    }

    var yHigh = y.$high;
    var yLow = y.$low;
    if (y.$high < 0) {
        s *= -1;
        yHigh = -yHigh;
        if (yLow !== 0) {
            yHigh--;
            yLow = 4294967296 - yLow;
        }
    }

    var high = 0, low = 0, n = 0;
    while (yHigh < 2147483648 && ((xHigh > yHigh) || (xHigh === yHigh && xLow > yLow))) {
        yHigh = (yHigh << 1 | yLow >>> 31) >>> 0;
        yLow = (yLow << 1) >>> 0;
        n++;
    }
    for (var i = 0; i <= n; i++) {
        high = high << 1 | low >>> 31;
        low = (low << 1) >>> 0;
        if ((xHigh > yHigh) || (xHigh === yHigh && xLow >= yLow)) {
            xHigh = xHigh - yHigh;
            xLow = xLow - yLow;
            if (xLow < 0) {
                xHigh--;
                xLow += 4294967296;
            }
            low++;
            if (low === 4294967296) {
                high++;
                low = 0;
            }
        }
        yLow = (yLow >>> 1 | yHigh << (32 - 1)) >>> 0;
        yHigh = yHigh >>> 1;
    }

    if (returnRemainder) {
        return new x.constructor(xHigh * rs, xLow * rs);
    }
    return new x.constructor(high * s, low * s);
};

var $divComplex = (n, d) => {
    var ninf = n.$real === Infinity || n.$real === -Infinity || n.$imag === Infinity || n.$imag === -Infinity;
    var dinf = d.$real === Infinity || d.$real === -Infinity || d.$imag === Infinity || d.$imag === -Infinity;
    var nnan = !ninf && (n.$real !== n.$real || n.$imag !== n.$imag);
    var dnan = !dinf && (d.$real !== d.$real || d.$imag !== d.$imag);
    if (nnan || dnan) {
        return new n.constructor(NaN, NaN);
    }
    if (ninf && !dinf) {
        return new n.constructor(Infinity, Infinity);
    }
    if (!ninf && dinf) {
        return new n.constructor(0, 0);
    }
    if (d.$real === 0 && d.$imag === 0) {
        if (n.$real === 0 && n.$imag === 0) {
            return new n.constructor(NaN, NaN);
        }
        return new n.constructor(Infinity, Infinity);
    }
    var a = Math.abs(d.$real);
    var b = Math.abs(d.$imag);
    if (a <= b) {
        var ratio = d.$real / d.$imag;
        var denom = d.$real * ratio + d.$imag;
        return new n.constructor((n.$real * ratio + n.$imag) / denom, (n.$imag * ratio - n.$real) / denom);
    }
    var ratio = d.$imag / d.$real;
    var denom = d.$imag * ratio + d.$real;
    return new n.constructor((n.$imag * ratio + n.$real) / denom, (n.$imag - n.$real * ratio) / denom);
};
var $kindBool = 1;
var $kindInt = 2;
var $kindInt8 = 3;
var $kindInt16 = 4;
var $kindInt32 = 5;
var $kindInt64 = 6;
var $kindUint = 7;
var $kindUint8 = 8;
var $kindUint16 = 9;
var $kindUint32 = 10;
var $kindUint64 = 11;
var $kindUintptr = 12;
var $kindFloat32 = 13;
var $kindFloat64 = 14;
var $kindComplex64 = 15;
var $kindComplex128 = 16;
var $kindArray = 17;
var $kindChan = 18;
var $kindFunc = 19;
var $kindInterface = 20;
var $kindMap = 21;
var $kindPtr = 22;
var $kindSlice = 23;
var $kindString = 24;
var $kindStruct = 25;
var $kindUnsafePointer = 26;

var $methodSynthesizers = [];
var $addMethodSynthesizer = f => {
    if ($methodSynthesizers === null) {
        f();
        return;
    }
    $methodSynthesizers.push(f);
};
var $synthesizeMethods = () => {
    $methodSynthesizers.forEach(f => { f(); });
    $methodSynthesizers = null;
};

var $ifaceKeyFor = x => {
    if (x === $ifaceNil) {
        return 'nil';
    }
    var c = x.constructor;
    return c.string + '$' + c.keyFor(x.$val);
};

var $identity = x => { return x; };

var $typeIDCounter = 0;

var $idKey = x => {
    if (x.$id === undefined) {
        $idCounter++;
        x.$id = $idCounter;
    }
    return String(x.$id);
};

// Creates constructor functions for array pointer types. Returns a new function
// instace each time to make sure each type is independent of the other.
var $arrayPtrCtor = () => {
    return function (array) {
        this.$get = () => { return array; };
        this.$set = function (v) { typ.copy(this, v); };
        this.$val = array;
    };
}

var $newType = (size, kind, string, named, pkg, exported, constructor) => {
    var typ;
    switch (kind) {
        case $kindBool:
        case $kindInt:
        case $kindInt8:
        case $kindInt16:
        case $kindInt32:
        case $kindUint:
        case $kindUint8:
        case $kindUint16:
        case $kindUint32:
        case $kindUintptr:
        case $kindUnsafePointer:
            typ = function (v) { this.$val = v; };
            typ.wrapped = true;
            typ.keyFor = $identity;
            break;

        case $kindString:
            typ = function (v) { this.$val = v; };
            typ.wrapped = true;
            typ.keyFor = x => { return "$" + x; };
            break;

        case $kindFloat32:
        case $kindFloat64:
            typ = function (v) { this.$val = v; };
            typ.wrapped = true;
            typ.keyFor = x => { return $floatKey(x); };
            break;

        case $kindInt64:
            typ = function (high, low) {
                this.$high = (high + Math.floor(Math.ceil(low) / 4294967296)) >> 0;
                this.$low = low >>> 0;
                this.$val = this;
            };
            typ.keyFor = x => { return x.$high + "$" + x.$low; };
            break;

        case $kindUint64:
            typ = function (high, low) {
                this.$high = (high + Math.floor(Math.ceil(low) / 4294967296)) >>> 0;
                this.$low = low >>> 0;
                this.$val = this;
            };
            typ.keyFor = x => { return x.$high + "$" + x.$low; };
            break;

        case $kindComplex64:
            typ = function (real, imag) {
                this.$real = $fround(real);
                this.$imag = $fround(imag);
                this.$val = this;
            };
            typ.keyFor = x => { return x.$real + "$" + x.$imag; };
            break;

        case $kindComplex128:
            typ = function (real, imag) {
                this.$real = real;
                this.$imag = imag;
                this.$val = this;
            };
            typ.keyFor = x => { return x.$real + "$" + x.$imag; };
            break;

        case $kindArray:
            typ = function (v) { this.$val = v; };
            typ.wrapped = true;
            typ.ptr = $newType(4, $kindPtr, "*" + string, false, "", false, $arrayPtrCtor());
            typ.init = (elem, len) => {
                typ.elem = elem;
                typ.len = len;
                typ.comparable = elem.comparable;
                typ.keyFor = x => {
                    return Array.prototype.join.call($mapArray(x, e => {
                        return String(elem.keyFor(e)).replace(/\\/g, "\\\\").replace(/\$/g, "\\$");
                    }), "$");
                };
                typ.copy = (dst, src) => {
                    $copyArray(dst, src, 0, 0, src.length, elem);
                };
                typ.ptr.init(typ);
                Object.defineProperty(typ.ptr.nil, "nilCheck", { get: $throwNilPointerError });
            };
            break;

        case $kindChan:
            typ = function (v) { this.$val = v; };
            typ.wrapped = true;
            typ.keyFor = $idKey;
            typ.init = (elem, sendOnly, recvOnly) => {
                typ.elem = elem;
                typ.sendOnly = sendOnly;
                typ.recvOnly = recvOnly;
            };
            break;

        case $kindFunc:
            typ = function (v) { this.$val = v; };
            typ.wrapped = true;
            typ.init = (params, results, variadic) => {
                typ.params = params;
                typ.results = results;
                typ.variadic = variadic;
                typ.comparable = false;
            };
            break;

        case $kindInterface:
            typ = { implementedBy: {}, missingMethodFor: {} };
            typ.keyFor = $ifaceKeyFor;
            typ.init = methods => {
                typ.methods = methods;
                methods.forEach(m => {
                    $ifaceNil[m.prop] = $throwNilPointerError;
                });
            };
            break;

        case $kindMap:
            typ = function (v) { this.$val = v; };
            typ.wrapped = true;
            typ.init = (key, elem) => {
                typ.key = key;
                typ.elem = elem;
                typ.comparable = false;
            };
            break;

        case $kindPtr:
            typ = constructor || function (getter, setter, target) {
                this.$get = getter;
                this.$set = setter;
                this.$target = target;
                this.$val = this;
            };
            typ.keyFor = $idKey;
            typ.init = elem => {
                typ.elem = elem;
                typ.wrapped = (elem.kind === $kindArray);
                typ.nil = new typ($throwNilPointerError, $throwNilPointerError);
            };
            break;

        case $kindSlice:
            typ = function (array) {
                if (array.constructor !== typ.nativeArray) {
                    array = new typ.nativeArray(array);
                }
                this.$array = array;
                this.$offset = 0;
                this.$length = array.length;
                this.$capacity = array.length;
                this.$val = this;
            };
            typ.init = elem => {
                typ.elem = elem;
                typ.comparable = false;
                typ.nativeArray = $nativeArray(elem.kind);
                typ.nil = new typ([]);
            };
            break;

        case $kindStruct:
            typ = function (v) { this.$val = v; };
            typ.wrapped = true;
            typ.ptr = $newType(4, $kindPtr, "*" + string, false, pkg, exported, constructor);
            typ.ptr.elem = typ;
            typ.ptr.prototype.$get = function () { return this; };
            typ.ptr.prototype.$set = function (v) { typ.copy(this, v); };
            typ.init = (pkgPath, fields) => {
                typ.pkgPath = pkgPath;
                typ.fields = fields;
                fields.forEach(f => {
                    if (!f.typ.comparable) {
                        typ.comparable = false;
                    }
                });
                typ.keyFor = x => {
                    var val = x.$val;
                    return $mapArray(fields, f => {
                        return String(f.typ.keyFor(val[f.prop])).replace(/\\/g, "\\\\").replace(/\$/g, "\\$");
                    }).join("$");
                };
                typ.copy = (dst, src) => {
                    for (var i = 0; i < fields.length; i++) {
                        var f = fields[i];
                        switch (f.typ.kind) {
                            case $kindArray:
                            case $kindStruct:
                                f.typ.copy(dst[f.prop], src[f.prop]);
                                continue;
                            default:
                                dst[f.prop] = src[f.prop];
                                continue;
                        }
                    }
                };
                /* nil value */
                var properties = {};
                fields.forEach(f => {
                    properties[f.prop] = { get: $throwNilPointerError, set: $throwNilPointerError };
                });
                typ.ptr.nil = Object.create(constructor.prototype, properties);
                typ.ptr.nil.$val = typ.ptr.nil;
                /* methods for embedded fields */
                $addMethodSynthesizer(() => {
                    var synthesizeMethod = (target, m, f) => {
                        if (target.prototype[m.prop] !== undefined) { return; }
                        target.prototype[m.prop] = function(...args) {
                            var v = this.$val[f.prop];
                            if (f.typ === $jsObjectPtr) {
                                v = new $jsObjectPtr(v);
                            }
                            if (v.$val === undefined) {
                                v = new f.typ(v);
                            }
                            return v[m.prop](...args);
                        };
                    };
                    fields.forEach(f => {
                        if (f.embedded) {
                            $methodSet(f.typ).forEach(m => {
                                synthesizeMethod(typ, m, f);
                                synthesizeMethod(typ.ptr, m, f);
                            });
                            $methodSet($ptrType(f.typ)).forEach(m => {
                                synthesizeMethod(typ.ptr, m, f);
                            });
                        }
                    });
                });
            };
            break;

        default:
            $panic(new $String("invalid kind: " + kind));
    }

    switch (kind) {
        case $kindBool:
        case $kindMap:
            typ.zero = () => { return false; };
            break;

        case $kindInt:
        case $kindInt8:
        case $kindInt16:
        case $kindInt32:
        case $kindUint:
        case $kindUint8:
        case $kindUint16:
        case $kindUint32:
        case $kindUintptr:
        case $kindUnsafePointer:
        case $kindFloat32:
        case $kindFloat64:
            typ.zero = () => { return 0; };
            break;

        case $kindString:
            typ.zero = () => { return ""; };
            break;

        case $kindInt64:
        case $kindUint64:
        case $kindComplex64:
        case $kindComplex128:
            var zero = new typ(0, 0);
            typ.zero = () => { return zero; };
            break;

        case $kindPtr:
        case $kindSlice:
            typ.zero = () => { return typ.nil; };
            break;

        case $kindChan:
            typ.zero = () => { return $chanNil; };
            break;

        case $kindFunc:
            typ.zero = () => { return $throwNilPointerError; };
            break;

        case $kindInterface:
            typ.zero = () => { return $ifaceNil; };
            break;

        case $kindArray:
            typ.zero = () => {
                var arrayClass = $nativeArray(typ.elem.kind);
                if (arrayClass !== Array) {
                    return new arrayClass(typ.len);
                }
                var array = new Array(typ.len);
                for (var i = 0; i < typ.len; i++) {
                    array[i] = typ.elem.zero();
                }
                return array;
            };
            break;

        case $kindStruct:
            typ.zero = () => { return new typ.ptr(); };
            break;

        default:
            $panic(new $String("invalid kind: " + kind));
    }

    typ.id = $typeIDCounter;
    $typeIDCounter++;
    typ.size = size;
    typ.kind = kind;
    typ.string = string;
    typ.named = named;
    typ.pkg = pkg;
    typ.exported = exported;
    typ.methods = [];
    typ.methodSetCache = null;
    typ.comparable = true;
    return typ;
};

var $methodSet = typ => {
    if (typ.methodSetCache !== null) {
        return typ.methodSetCache;
    }
    var base = {};

    var isPtr = (typ.kind === $kindPtr);
    if (isPtr && typ.elem.kind === $kindInterface) {
        typ.methodSetCache = [];
        return [];
    }

    var current = [{ typ: isPtr ? typ.elem : typ, indirect: isPtr }];

    var seen = {};

    while (current.length > 0) {
        var next = [];
        var mset = [];

        current.forEach(e => {
            if (seen[e.typ.string]) {
                return;
            }
            seen[e.typ.string] = true;

            if (e.typ.named) {
                mset = mset.concat(e.typ.methods);
                if (e.indirect) {
                    mset = mset.concat($ptrType(e.typ).methods);
                }
            }

            switch (e.typ.kind) {
                case $kindStruct:
                    e.typ.fields.forEach(f => {
                        if (f.embedded) {
                            var fTyp = f.typ;
                            var fIsPtr = (fTyp.kind === $kindPtr);
                            next.push({ typ: fIsPtr ? fTyp.elem : fTyp, indirect: e.indirect || fIsPtr });
                        }
                    });
                    break;

                case $kindInterface:
                    mset = mset.concat(e.typ.methods);
                    break;
            }
        });

        mset.forEach(m => {
            if (base[m.name] === undefined) {
                base[m.name] = m;
            }
        });

        current = next;
    }

    typ.methodSetCache = [];
    Object.keys(base).sort().forEach(name => {
        typ.methodSetCache.push(base[name]);
    });
    return typ.methodSetCache;
};

var $Bool = $newType(1, $kindBool, "bool", true, "", false, null);
var $Int = $newType(4, $kindInt, "int", true, "", false, null);
var $Int8 = $newType(1, $kindInt8, "int8", true, "", false, null);
var $Int16 = $newType(2, $kindInt16, "int16", true, "", false, null);
var $Int32 = $newType(4, $kindInt32, "int32", true, "", false, null);
var $Int64 = $newType(8, $kindInt64, "int64", true, "", false, null);
var $Uint = $newType(4, $kindUint, "uint", true, "", false, null);
var $Uint8 = $newType(1, $kindUint8, "uint8", true, "", false, null);
var $Uint16 = $newType(2, $kindUint16, "uint16", true, "", false, null);
var $Uint32 = $newType(4, $kindUint32, "uint32", true, "", false, null);
var $Uint64 = $newType(8, $kindUint64, "uint64", true, "", false, null);
var $Uintptr = $newType(4, $kindUintptr, "uintptr", true, "", false, null);
var $Float32 = $newType(4, $kindFloat32, "float32", true, "", false, null);
var $Float64 = $newType(8, $kindFloat64, "float64", true, "", false, null);
var $Complex64 = $newType(8, $kindComplex64, "complex64", true, "", false, null);
var $Complex128 = $newType(16, $kindComplex128, "complex128", true, "", false, null);
var $String = $newType(8, $kindString, "string", true, "", false, null);
var $UnsafePointer = $newType(4, $kindUnsafePointer, "unsafe.Pointer", true, "unsafe", false, null);

var $nativeArray = elemKind => {
    switch (elemKind) {
        case $kindInt:
            return Int32Array;
        case $kindInt8:
            return Int8Array;
        case $kindInt16:
            return Int16Array;
        case $kindInt32:
            return Int32Array;
        case $kindUint:
            return Uint32Array;
        case $kindUint8:
            return Uint8Array;
        case $kindUint16:
            return Uint16Array;
        case $kindUint32:
            return Uint32Array;
        case $kindUintptr:
            return Uint32Array;
        case $kindFloat32:
            return Float32Array;
        case $kindFloat64:
            return Float64Array;
        default:
            return Array;
    }
};
var $toNativeArray = (elemKind, array) => {
    var nativeArray = $nativeArray(elemKind);
    if (nativeArray === Array) {
        return array;
    }
    return new nativeArray(array);
};
var $arrayTypes = {};
var $arrayType = (elem, len) => {
    var typeKey = elem.id + "$" + len;
    var typ = $arrayTypes[typeKey];
    if (typ === undefined) {
        typ = $newType(elem.size * len, $kindArray, "[" + len + "]" + elem.string, false, "", false, null);
        $arrayTypes[typeKey] = typ;
        typ.init(elem, len);
    }
    return typ;
};

var $chanType = (elem, sendOnly, recvOnly) => {
    var string = (recvOnly ? "<-" : "") + "chan" + (sendOnly ? "<- " : " ");
    if (!sendOnly && !recvOnly && (elem.string[0] == "<")) {
        string += "(" + elem.string + ")";
    } else {
        string += elem.string;
    }
    var field = sendOnly ? "SendChan" : (recvOnly ? "RecvChan" : "Chan");
    var typ = elem[field];
    if (typ === undefined) {
        typ = $newType(4, $kindChan, string, false, "", false, null);
        elem[field] = typ;
        typ.init(elem, sendOnly, recvOnly);
    }
    return typ;
};
var $Chan = function (elem, capacity) {
    if (capacity < 0 || capacity > 2147483647) {
        $throwRuntimeError("makechan: size out of range");
    }
    this.$elem = elem;
    this.$capacity = capacity;
    this.$buffer = [];
    this.$sendQueue = [];
    this.$recvQueue = [];
    this.$closed = false;
};
var $chanNil = new $Chan(null, 0);
$chanNil.$sendQueue = $chanNil.$recvQueue = { length: 0, push() { }, shift() { return undefined; }, indexOf() { return -1; } };

var $funcTypes = {};
var $funcType = (params, results, variadic) => {
    var typeKey = $mapArray(params, p => { return p.id; }).join(",") + "$" + $mapArray(results, r => { return r.id; }).join(",") + "$" + variadic;
    var typ = $funcTypes[typeKey];
    if (typ === undefined) {
        var paramTypes = $mapArray(params, p => { return p.string; });
        if (variadic) {
            paramTypes[paramTypes.length - 1] = "..." + paramTypes[paramTypes.length - 1].substr(2);
        }
        var string = "func(" + paramTypes.join(", ") + ")";
        if (results.length === 1) {
            string += " " + results[0].string;
        } else if (results.length > 1) {
            string += " (" + $mapArray(results, r => { return r.string; }).join(", ") + ")";
        }
        typ = $newType(4, $kindFunc, string, false, "", false, null);
        $funcTypes[typeKey] = typ;
        typ.init(params, results, variadic);
    }
    return typ;
};

var $interfaceTypes = {};
var $interfaceType = methods => {
    var typeKey = $mapArray(methods, m => { return m.pkg + "," + m.name + "," + m.typ.id; }).join("$");
    var typ = $interfaceTypes[typeKey];
    if (typ === undefined) {
        var string = "interface {}";
        if (methods.length !== 0) {
            string = "interface { " + $mapArray(methods, m => {
                return (m.pkg !== "" ? m.pkg + "." : "") + m.name + m.typ.string.substr(4);
            }).join("; ") + " }";
        }
        typ = $newType(8, $kindInterface, string, false, "", false, null);
        $interfaceTypes[typeKey] = typ;
        typ.init(methods);
    }
    return typ;
};
var $emptyInterface = $interfaceType([]);
var $ifaceNil = {};
var $error = $newType(8, $kindInterface, "error", true, "", false, null);
$error.init([{ prop: "Error", name: "Error", pkg: "", typ: $funcType([], [$String], false) }]);

var $mapTypes = {};
var $mapType = (key, elem) => {
    var typeKey = key.id + "$" + elem.id;
    var typ = $mapTypes[typeKey];
    if (typ === undefined) {
        typ = $newType(4, $kindMap, "map[" + key.string + "]" + elem.string, false, "", false, null);
        $mapTypes[typeKey] = typ;
        typ.init(key, elem);
    }
    return typ;
};
var $makeMap = (keyForFunc, entries) => {
    var m = new Map();
    for (var i = 0; i < entries.length; i++) {
        var e = entries[i];
        m.set(keyForFunc(e.k), e);
    }
    return m;
};

var $ptrType = elem => {
    var typ = elem.ptr;
    if (typ === undefined) {
        typ = $newType(4, $kindPtr, "*" + elem.string, false, "", elem.exported, null);
        elem.ptr = typ;
        typ.init(elem);
    }
    return typ;
};

var $newDataPointer = (data, constructor) => {
    if (constructor.elem.kind === $kindStruct) {
        return data;
    }
    return new constructor(() => { return data; }, v => { data = v; });
};

var $indexPtr = (array, index, constructor) => {
    if (array.buffer) {
        // Pointers to the same underlying ArrayBuffer share cache.
        var cache = array.buffer.$ptr = array.buffer.$ptr || {};
        // Pointers of different primitive types are non-comparable and stored in different caches.
        var typeCache = cache[array.name] = cache[array.name] || {};
        var cacheIdx = array.BYTES_PER_ELEMENT * index + array.byteOffset;
        return typeCache[cacheIdx] || (typeCache[cacheIdx] = new constructor(() => { return array[index]; }, v => { array[index] = v; }));
    } else {
        array.$ptr = array.$ptr || {};
        return array.$ptr[index] || (array.$ptr[index] = new constructor(() => { return array[index]; }, v => { array[index] = v; }));
    }
};

var $sliceType = elem => {
    var typ = elem.slice;
    if (typ === undefined) {
        typ = $newType(12, $kindSlice, "[]" + elem.string, false, "", false, null);
        elem.slice = typ;
        typ.init(elem);
    }
    return typ;
};
var $makeSlice = (typ, length, capacity = length) => {
    if (length < 0 || length > 2147483647) {
        $throwRuntimeError("makeslice: len out of range");
    }
    if (capacity < 0 || capacity < length || capacity > 2147483647) {
        $throwRuntimeError("makeslice: cap out of range");
    }
    var array = new typ.nativeArray(capacity);
    if (typ.nativeArray === Array) {
        for (var i = 0; i < capacity; i++) {
            array[i] = typ.elem.zero();
        }
    }
    var slice = new typ(array);
    slice.$length = length;
    return slice;
};

var $structTypes = {};
var $structType = (pkgPath, fields) => {
    var typeKey = $mapArray(fields, f => { return f.name + "," + f.typ.id + "," + f.tag; }).join("$");
    var typ = $structTypes[typeKey];
    if (typ === undefined) {
        var string = "struct { " + $mapArray(fields, f => {
            var str = f.typ.string + (f.tag !== "" ? (" \"" + f.tag.replace(/\\/g, "\\\\").replace(/"/g, "\\\"") + "\"") : "");
            if (f.embedded) {
                return str;
            }
            return f.name + " " + str;
        }).join("; ") + " }";
        if (fields.length === 0) {
            string = "struct {}";
        }
        typ = $newType(0, $kindStruct, string, false, "", false, function(...args) {
            this.$val = this;
            for (var i = 0; i < fields.length; i++) {
                var f = fields[i];
                if (f.name == '_') {
                    continue;
                }
                var arg = args[i];
                this[f.prop] = arg !== undefined ? arg : f.typ.zero();
            }
        });
        $structTypes[typeKey] = typ;
        typ.init(pkgPath, fields);
    }
    return typ;
};

var $assertType = (value, type, returnTuple) => {
    var isInterface = (type.kind === $kindInterface), ok, missingMethod = "";
    if (value === $ifaceNil) {
        ok = false;
    } else if (!isInterface) {
        ok = value.constructor === type;
    } else {
        var valueTypeString = value.constructor.string;
        ok = type.implementedBy[valueTypeString];
        if (ok === undefined) {
            ok = true;
            var valueMethodSet = $methodSet(value.constructor);
            var interfaceMethods = type.methods;
            for (var i = 0; i < interfaceMethods.length; i++) {
                var tm = interfaceMethods[i];
                var found = false;
                for (var j = 0; j < valueMethodSet.length; j++) {
                    var vm = valueMethodSet[j];
                    if (vm.name === tm.name && vm.pkg === tm.pkg && vm.typ === tm.typ) {
                        found = true;
                        break;
                    }
                }
                if (!found) {
                    ok = false;
                    type.missingMethodFor[valueTypeString] = tm.name;
                    break;
                }
            }
            type.implementedBy[valueTypeString] = ok;
        }
        if (!ok) {
            missingMethod = type.missingMethodFor[valueTypeString];
        }
    }

    if (!ok) {
        if (returnTuple) {
            return [type.zero(), false];
        }
        $panic(new $packages["runtime"].TypeAssertionError.ptr(
            $packages["runtime"]._type.ptr.nil,
            (value === $ifaceNil ? $packages["runtime"]._type.ptr.nil : new $packages["runtime"]._type.ptr(value.constructor.string)),
            new $packages["runtime"]._type.ptr(type.string),
            missingMethod));
    }

    if (!isInterface) {
        value = value.$val;
    }
    if (type === $jsObjectPtr) {
        value = value.object;
    }
    return returnTuple ? [value, true] : value;
};
var $stackDepthOffset = 0;
var $getStackDepth = () => {
    var err = new Error();
    if (err.stack === undefined) {
        return undefined;
    }
    return $stackDepthOffset + err.stack.split("\n").length;
};

var $panicStackDepth = null, $panicValue;
var $callDeferred = (deferred, jsErr, fromPanic) => {
    if (!fromPanic && deferred !== null && $curGoroutine.deferStack.indexOf(deferred) == -1) {
        throw jsErr;
    }
    if (jsErr !== null) {
        var newErr = null;
        try {
            $panic(new $jsErrorPtr(jsErr));
        } catch (err) {
            newErr = err;
        }
        $callDeferred(deferred, newErr);
        return;
    }
    if ($curGoroutine.asleep) {
        return;
    }

    $stackDepthOffset--;
    var outerPanicStackDepth = $panicStackDepth;
    var outerPanicValue = $panicValue;

    var localPanicValue = $curGoroutine.panicStack.pop();
    if (localPanicValue !== undefined) {
        $panicStackDepth = $getStackDepth();
        $panicValue = localPanicValue;
    }

    try {
        while (true) {
            if (deferred === null) {
                deferred = $curGoroutine.deferStack[$curGoroutine.deferStack.length - 1];
                if (deferred === undefined) {
                    /* The panic reached the top of the stack. Clear it and throw it as a JavaScript error. */
                    $panicStackDepth = null;
                    if (localPanicValue.Object instanceof Error) {
                        throw localPanicValue.Object;
                    }
                    var msg;
                    if (localPanicValue.constructor === $String) {
                        msg = localPanicValue.$val;
                    } else if (localPanicValue.Error !== undefined) {
                        msg = localPanicValue.Error();
                    } else if (localPanicValue.String !== undefined) {
                        msg = localPanicValue.String();
                    } else {
                        msg = localPanicValue;
                    }
                    throw new Error(msg);
                }
            }
            var call = deferred.pop();
            if (call === undefined) {
                $curGoroutine.deferStack.pop();
                if (localPanicValue !== undefined) {
                    deferred = null;
                    continue;
                }
                return;
            }
            var r = call[0].apply(call[2], call[1]);
            if (r && r.$blk !== undefined) {
                deferred.push([r.$blk, [], r]);
                if (fromPanic) {
                    throw null;
                }
                return;
            }

            if (localPanicValue !== undefined && $panicStackDepth === null) {
                /* error was recovered */
                if (fromPanic) {
                    throw null;
                }
                return;
            }
        }
    } catch (e) {
        // Deferred function threw a JavaScript exception or tries to unwind stack
        // to the point where a panic was handled.
        if (fromPanic) {
            // Re-throw the exception to reach deferral execution call at the end
            // of the function.
            throw e;
        }
        // We are at the end of the function, handle the error or re-throw to
        // continue unwinding if necessary, or simply stop unwinding if we got far
        // enough.
        $callDeferred(deferred, e, fromPanic);
    } finally {
        if (localPanicValue !== undefined) {
            if ($panicStackDepth !== null) {
                $curGoroutine.panicStack.push(localPanicValue);
            }
            $panicStackDepth = outerPanicStackDepth;
            $panicValue = outerPanicValue;
        }
        $stackDepthOffset++;
    }
};

var $panic = value => {
    $curGoroutine.panicStack.push(value);
    $callDeferred(null, null, true);
};
var $recover = () => {
    if ($panicStackDepth === null || ($panicStackDepth !== undefined && $panicStackDepth !== $getStackDepth() - 2)) {
        return $ifaceNil;
    }
    $panicStackDepth = null;
    return $panicValue;
};
var $throw = err => { throw err; };

var $noGoroutine = { asleep: false, exit: false, deferStack: [], panicStack: [] };
var $curGoroutine = $noGoroutine, $totalGoroutines = 0, $awakeGoroutines = 0, $checkForDeadlock = true, $exportedFunctions = 0;
var $mainFinished = false;
var $go = (fun, args) => {
    $totalGoroutines++;
    $awakeGoroutines++;
    var $goroutine = () => {
        try {
            $curGoroutine = $goroutine;
            var r = fun(...args);
            if (r && r.$blk !== undefined) {
                fun = () => { return r.$blk(); };
                args = [];
                return;
            }
            $goroutine.exit = true;
        } catch (err) {
            if (!$goroutine.exit) {
                throw err;
            }
        } finally {
            $curGoroutine = $noGoroutine;
            if ($goroutine.exit) { /* also set by runtime.Goexit() */
                $totalGoroutines--;
                $goroutine.asleep = true;
            }
            if ($goroutine.asleep) {
                $awakeGoroutines--;
                if (!$mainFinished && $awakeGoroutines === 0 && $checkForDeadlock && $exportedFunctions === 0) {
                    console.error("fatal error: all goroutines are asleep - deadlock!");
                    if ($global.process !== undefined) {
                        $global.process.exit(2);
                    }
                }
            }
        }
    };
    $goroutine.asleep = false;
    $goroutine.exit = false;
    $goroutine.deferStack = [];
    $goroutine.panicStack = [];
    $schedule($goroutine);
};

var $scheduled = [];
var $runScheduled = () => {
    // For nested setTimeout calls browsers enforce 4ms minimum delay. We minimize
    // the effect of this penalty by queueing the timer preemptively before we run
    // the goroutines, and later cancelling it if it turns out unneeded. See:
    // https://developer.mozilla.org/en-US/docs/Web/API/setTimeout#nested_timeouts
    var nextRun = setTimeout($runScheduled);
    try {
        var start = Date.now();
        var r;
        while ((r = $scheduled.shift()) !== undefined) {
            r();
            // We need to interrupt this loop in order to allow the event loop to
            // process timers, IO, etc. However, invoking scheduling through
            // setTimeout is ~1000 times more expensive, so we amortize this cost by
            // looping until the 4ms minimal delay has elapsed (assuming there are
            // scheduled goroutines to run), and then yield to the event loop.
            var elapsed = Date.now() - start;
            if (elapsed > 4 || elapsed < 0) { break; }
        }
    } finally {
        if ($scheduled.length == 0) {
            // Cancel scheduling pass if there's nothing to run.
            clearTimeout(nextRun);
        }
    }
};

var $schedule = goroutine => {
    if (goroutine.asleep) {
        goroutine.asleep = false;
        $awakeGoroutines++;
    }
    $scheduled.push(goroutine);
    if ($curGoroutine === $noGoroutine) {
        $runScheduled();
    }
};

var $setTimeout = (f, t) => {
    $awakeGoroutines++;
    return setTimeout(() => {
        $awakeGoroutines--;
        f();
    }, t);
};

var $block = () => {
    if ($curGoroutine === $noGoroutine) {
        $throwRuntimeError("cannot block in JavaScript callback, fix by wrapping code in goroutine");
    }
    $curGoroutine.asleep = true;
};

var $restore = (context, params) => {
    if (context !== undefined && context.$blk !== undefined) {
        return context;
    }
    return params;
}

var $send = (chan, value) => {
    if (chan.$closed) {
        $throwRuntimeError("send on closed channel");
    }
    var queuedRecv = chan.$recvQueue.shift();
    if (queuedRecv !== undefined) {
        queuedRecv([value, true]);
        return;
    }
    if (chan.$buffer.length < chan.$capacity) {
        chan.$buffer.push(value);
        return;
    }

    var thisGoroutine = $curGoroutine;
    var closedDuringSend;
    chan.$sendQueue.push(closed => {
        closedDuringSend = closed;
        $schedule(thisGoroutine);
        return value;
    });
    $block();
    return {
        $blk() {
            if (closedDuringSend) {
                $throwRuntimeError("send on closed channel");
            }
        }
    };
};
var $recv = chan => {
    var queuedSend = chan.$sendQueue.shift();
    if (queuedSend !== undefined) {
        chan.$buffer.push(queuedSend(false));
    }
    var bufferedValue = chan.$buffer.shift();
    if (bufferedValue !== undefined) {
        return [bufferedValue, true];
    }
    if (chan.$closed) {
        return [chan.$elem.zero(), false];
    }

    var thisGoroutine = $curGoroutine;
    var f = { $blk() { return this.value; } };
    var queueEntry = v => {
        f.value = v;
        $schedule(thisGoroutine);
    };
    chan.$recvQueue.push(queueEntry);
    $block();
    return f;
};
var $close = chan => {
    if (chan.$closed) {
        $throwRuntimeError("close of closed channel");
    }
    chan.$closed = true;
    while (true) {
        var queuedSend = chan.$sendQueue.shift();
        if (queuedSend === undefined) {
            break;
        }
        queuedSend(true); /* will panic */
    }
    while (true) {
        var queuedRecv = chan.$recvQueue.shift();
        if (queuedRecv === undefined) {
            break;
        }
        queuedRecv([chan.$elem.zero(), false]);
    }
};
var $select = comms => {
    var ready = [];
    var selection = -1;
    for (var i = 0; i < comms.length; i++) {
        var comm = comms[i];
        var chan = comm[0];
        switch (comm.length) {
            case 0: /* default */
                selection = i;
                break;
            case 1: /* recv */
                if (chan.$sendQueue.length !== 0 || chan.$buffer.length !== 0 || chan.$closed) {
                    ready.push(i);
                }
                break;
            case 2: /* send */
                if (chan.$closed) {
                    $throwRuntimeError("send on closed channel");
                }
                if (chan.$recvQueue.length !== 0 || chan.$buffer.length < chan.$capacity) {
                    ready.push(i);
                }
                break;
        }
    }

    if (ready.length !== 0) {
        selection = ready[Math.floor(Math.random() * ready.length)];
    }
    if (selection !== -1) {
        var comm = comms[selection];
        switch (comm.length) {
            case 0: /* default */
                return [selection];
            case 1: /* recv */
                return [selection, $recv(comm[0])];
            case 2: /* send */
                $send(comm[0], comm[1]);
                return [selection];
        }
    }

    var entries = [];
    var thisGoroutine = $curGoroutine;
    var f = { $blk() { return this.selection; } };
    var removeFromQueues = () => {
        for (var i = 0; i < entries.length; i++) {
            var entry = entries[i];
            var queue = entry[0];
            var index = queue.indexOf(entry[1]);
            if (index !== -1) {
                queue.splice(index, 1);
            }
        }
    };
    for (var i = 0; i < comms.length; i++) {
        (i => {
            var comm = comms[i];
            switch (comm.length) {
                case 1: /* recv */
                    var queueEntry = value => {
                        f.selection = [i, value];
                        removeFromQueues();
                        $schedule(thisGoroutine);
                    };
                    entries.push([comm[0].$recvQueue, queueEntry]);
                    comm[0].$recvQueue.push(queueEntry);
                    break;
                case 2: /* send */
                    var queueEntry = () => {
                        if (comm[0].$closed) {
                            $throwRuntimeError("send on closed channel");
                        }
                        f.selection = [i];
                        removeFromQueues();
                        $schedule(thisGoroutine);
                        return comm[1];
                    };
                    entries.push([comm[0].$sendQueue, queueEntry]);
                    comm[0].$sendQueue.push(queueEntry);
                    break;
            }
        })(i);
    }
    $block();
    return f;
};
var $jsObjectPtr, $jsErrorPtr;

var $needsExternalization = t => {
    switch (t.kind) {
        case $kindBool:
        case $kindInt:
        case $kindInt8:
        case $kindInt16:
        case $kindInt32:
        case $kindUint:
        case $kindUint8:
        case $kindUint16:
        case $kindUint32:
        case $kindUintptr:
        case $kindFloat32:
        case $kindFloat64:
            return false;
        default:
            return t !== $jsObjectPtr;
    }
};

var $externalize = (v, t, makeWrapper) => {
    if (t === $jsObjectPtr) {
        return v;
    }
    switch (t.kind) {
        case $kindBool:
        case $kindInt:
        case $kindInt8:
        case $kindInt16:
        case $kindInt32:
        case $kindUint:
        case $kindUint8:
        case $kindUint16:
        case $kindUint32:
        case $kindUintptr:
        case $kindFloat32:
        case $kindFloat64:
            return v;
        case $kindInt64:
        case $kindUint64:
            return $flatten64(v);
        case $kindArray:
            if ($needsExternalization(t.elem)) {
                return $mapArray(v, e => { return $externalize(e, t.elem, makeWrapper); });
            }
            return v;
        case $kindFunc:
            return $externalizeFunction(v, t, false, makeWrapper);
        case $kindInterface:
            if (v === $ifaceNil) {
                return null;
            }
            if (v.constructor === $jsObjectPtr) {
                return v.$val.object;
            }
            return $externalize(v.$val, v.constructor, makeWrapper);
        case $kindMap:
            if (v.keys === undefined) {
                return null;
            }
            var m = {};
            var keys = Array.from(v.keys());
            for (var i = 0; i < keys.length; i++) {
                var entry = v.get(keys[i]);
                m[$externalize(entry.k, t.key, makeWrapper)] = $externalize(entry.v, t.elem, makeWrapper);
            }
            return m;
        case $kindPtr:
            if (v === t.nil) {
                return null;
            }
            return $externalize(v.$get(), t.elem, makeWrapper);
        case $kindSlice:
            if (v === v.constructor.nil) {
                return null;
            }
            if ($needsExternalization(t.elem)) {
                return $mapArray($sliceToNativeArray(v), e => { return $externalize(e, t.elem, makeWrapper); });
            }
            return $sliceToNativeArray(v);
        case $kindString:
            if ($isASCII(v)) {
                return v;
            }
            var s = "", r;
            for (var i = 0; i < v.length; i += r[1]) {
                r = $decodeRune(v, i);
                var c = r[0];
                if (c > 0xFFFF) {
                    var h = Math.floor((c - 0x10000) / 0x400) + 0xD800;
                    var l = (c - 0x10000) % 0x400 + 0xDC00;
                    s += String.fromCharCode(h, l);
                    continue;
                }
                s += String.fromCharCode(c);
            }
            return s;
        case $kindStruct:
            var timePkg = $packages["time"];
            if (timePkg !== undefined && v.constructor === timePkg.Time.ptr) {
                var milli = $div64(v.UnixNano(), new $Int64(0, 1000000));
                return new Date($flatten64(milli));
            }

            var noJsObject = {};
            var searchJsObject = (v, t) => {
                if (t === $jsObjectPtr) {
                    return v;
                }
                switch (t.kind) {
                    case $kindPtr:
                        if (v === t.nil) {
                            return noJsObject;
                        }
                        return searchJsObject(v.$get(), t.elem);
                    case $kindStruct:
                        if (t.fields.length === 0) {
                            return noJsObject;
                        }
                        var f = t.fields[0];
                        return searchJsObject(v[f.prop], f.typ);
                    case $kindInterface:
                        return searchJsObject(v.$val, v.constructor);
                    default:
                        return noJsObject;
                }
            };
            var o = searchJsObject(v, t);
            if (o !== noJsObject) {
                return o;
            }

            if (makeWrapper !== undefined) {
                return makeWrapper(v);
            }

            o = {};
            for (var i = 0; i < t.fields.length; i++) {
                var f = t.fields[i];
                if (!f.exported) {
                    continue;
                }
                o[f.name] = $externalize(v[f.prop], f.typ, makeWrapper);
            }
            return o;
    }
    $throwRuntimeError("cannot externalize " + t.string);
};

var $externalizeFunction = (v, t, passThis, makeWrapper) => {
    if (v === $throwNilPointerError) {
        return null;
    }
    if (v.$externalizeWrapper === undefined) {
        $checkForDeadlock = false;
        v.$externalizeWrapper = function () {
            var args = [];
            for (var i = 0; i < t.params.length; i++) {
                if (t.variadic && i === t.params.length - 1) {
                    var vt = t.params[i].elem, varargs = [];
                    for (var j = i; j < arguments.length; j++) {
                        varargs.push($internalize(arguments[j], vt, makeWrapper));
                    }
                    args.push(new (t.params[i])(varargs));
                    break;
                }
                args.push($internalize(arguments[i], t.params[i], makeWrapper));
            }
            var result = v.apply(passThis ? this : undefined, args);
            switch (t.results.length) {
                case 0:
                    return;
                case 1:
                    return $externalize($copyIfRequired(result, t.results[0]), t.results[0], makeWrapper);
                default:
                    for (var i = 0; i < t.results.length; i++) {
                        result[i] = $externalize($copyIfRequired(result[i], t.results[i]), t.results[i], makeWrapper);
                    }
                    return result;
            }
        };
    }
    return v.$externalizeWrapper;
};

var $internalize = (v, t, recv, seen, makeWrapper) => {
    if (t === $jsObjectPtr) {
        return v;
    }
    if (t === $jsObjectPtr.elem) {
        $throwRuntimeError("cannot internalize js.Object, use *js.Object instead");
    }
    if (v && v.__internal_object__ !== undefined) {
        return $assertType(v.__internal_object__, t, false);
    }
    var timePkg = $packages["time"];
    if (timePkg !== undefined && t === timePkg.Time) {
        if (!(v !== null && v !== undefined && v.constructor === Date)) {
            $throwRuntimeError("cannot internalize time.Time from " + typeof v + ", must be Date");
        }
        return timePkg.Unix(new $Int64(0, 0), new $Int64(0, v.getTime() * 1000000));
    }

    // Cache for values we've already internalized in order to deal with circular
    // references.
    if (seen === undefined) { seen = new Map(); }
    if (!seen.has(t)) { seen.set(t, new Map()); }
    if (seen.get(t).has(v)) { return seen.get(t).get(v); }

    switch (t.kind) {
        case $kindBool:
            return !!v;
        case $kindInt:
            return parseInt(v);
        case $kindInt8:
            return parseInt(v) << 24 >> 24;
        case $kindInt16:
            return parseInt(v) << 16 >> 16;
        case $kindInt32:
            return parseInt(v) >> 0;
        case $kindUint:
            return parseInt(v);
        case $kindUint8:
            return parseInt(v) << 24 >>> 24;
        case $kindUint16:
            return parseInt(v) << 16 >>> 16;
        case $kindUint32:
        case $kindUintptr:
            return parseInt(v) >>> 0;
        case $kindInt64:
        case $kindUint64:
            return new t(0, v);
        case $kindFloat32:
        case $kindFloat64:
            return parseFloat(v);
        case $kindArray:
            if (v.length !== t.len) {
                $throwRuntimeError("got array with wrong size from JavaScript native");
            }
            return $mapArray(v, e => { return $internalize(e, t.elem, makeWrapper); });
        case $kindFunc:
            return function () {
                var args = [];
                for (var i = 0; i < t.params.length; i++) {
                    if (t.variadic && i === t.params.length - 1) {
                        var vt = t.params[i].elem, varargs = arguments[i];
                        for (var j = 0; j < varargs.$length; j++) {
                            args.push($externalize(varargs.$array[varargs.$offset + j], vt, makeWrapper));
                        }
                        break;
                    }
                    args.push($externalize(arguments[i], t.params[i], makeWrapper));
                }
                var result = v.apply(recv, args);
                switch (t.results.length) {
                    case 0:
                        return;
                    case 1:
                        return $internalize(result, t.results[0], makeWrapper);
                    default:
                        for (var i = 0; i < t.results.length; i++) {
                            result[i] = $internalize(result[i], t.results[i], makeWrapper);
                        }
                        return result;
                }
            };
        case $kindInterface:
            if (t.methods.length !== 0) {
                $throwRuntimeError("cannot internalize " + t.string);
            }
            if (v === null) {
                return $ifaceNil;
            }
            if (v === undefined) {
                return new $jsObjectPtr(undefined);
            }
            switch (v.constructor) {
                case Int8Array:
                    return new ($sliceType($Int8))(v);
                case Int16Array:
                    return new ($sliceType($Int16))(v);
                case Int32Array:
                    return new ($sliceType($Int))(v);
                case Uint8Array:
                    return new ($sliceType($Uint8))(v);
                case Uint16Array:
                    return new ($sliceType($Uint16))(v);
                case Uint32Array:
                    return new ($sliceType($Uint))(v);
                case Float32Array:
                    return new ($sliceType($Float32))(v);
                case Float64Array:
                    return new ($sliceType($Float64))(v);
                case Array:
                    return $internalize(v, $sliceType($emptyInterface), makeWrapper);
                case Boolean:
                    return new $Bool(!!v);
                case Date:
                    if (timePkg === undefined) {
                        /* time package is not present, internalize as &js.Object{Date} so it can be externalized into original Date. */
                        return new $jsObjectPtr(v);
                    }
                    return new timePkg.Time($internalize(v, timePkg.Time, makeWrapper));
                case ((() => { })).constructor: // is usually Function, but in Chrome extensions it is something else
                    var funcType = $funcType([$sliceType($emptyInterface)], [$jsObjectPtr], true);
                    return new funcType($internalize(v, funcType, makeWrapper));
                case Number:
                    return new $Float64(parseFloat(v));
                case String:
                    return new $String($internalize(v, $String, makeWrapper));
                default:
                    if ($global.Node && v instanceof $global.Node) {
                        return new $jsObjectPtr(v);
                    }
                    var mapType = $mapType($String, $emptyInterface);
                    return new mapType($internalize(v, mapType, recv, seen, makeWrapper));
            }
        case $kindMap:
            var m = new Map();
            seen.get(t).set(v, m);
            var keys = $keys(v);
            for (var i = 0; i < keys.length; i++) {
                var k = $internalize(keys[i], t.key, recv, seen, makeWrapper);
                m.set(t.key.keyFor(k), { k, v: $internalize(v[keys[i]], t.elem, recv, seen, makeWrapper) });
            }
            return m;
        case $kindPtr:
            if (t.elem.kind === $kindStruct) {
                return $internalize(v, t.elem, makeWrapper);
            }
        case $kindSlice:
            return new t($mapArray(v, e => { return $internalize(e, t.elem, makeWrapper); }));
        case $kindString:
            v = String(v);
            if ($isASCII(v)) {
                return v;
            }
            var s = "";
            var i = 0;
            while (i < v.length) {
                var h = v.charCodeAt(i);
                if (0xD800 <= h && h <= 0xDBFF) {
                    var l = v.charCodeAt(i + 1);
                    var c = (h - 0xD800) * 0x400 + l - 0xDC00 + 0x10000;
                    s += $encodeRune(c);
                    i += 2;
                    continue;
                }
                s += $encodeRune(h);
                i++;
            }
            return s;
        case $kindStruct:
            var noJsObject = {};
            var searchJsObject = t => {
                if (t === $jsObjectPtr) {
                    return v;
                }
                if (t === $jsObjectPtr.elem) {
                    $throwRuntimeError("cannot internalize js.Object, use *js.Object instead");
                }
                switch (t.kind) {
                    case $kindPtr:
                        return searchJsObject(t.elem);
                    case $kindStruct:
                        if (t.fields.length === 0) {
                            return noJsObject;
                        }
                        var f = t.fields[0];
                        var o = searchJsObject(f.typ);
                        if (o !== noJsObject) {
                            var n = new t.ptr();
                            n[f.prop] = o;
                            return n;
                        }
                        return noJsObject;
                    default:
                        return noJsObject;
                }
            };
            var o = searchJsObject(t);
            if (o !== noJsObject) {
                return o;
            }
            var n = new t.ptr();
            for (var i = 0; i < t.fields.length; i++) {
              var f = t.fields[i];
      
              if (!f.exported) {
                continue;
              }
              var jsProp = v[f.name];
      
              n[f.prop] = $internalize(jsProp, f.typ, recv, seen, makeWrapper);
            }
      
            return n;
    }
    $throwRuntimeError("cannot internalize " + t.string);
};

var $copyIfRequired = (v, typ) => {
    // interface values
    if (v && v.constructor && v.constructor.copy) {
        return new v.constructor($clone(v.$val, v.constructor))
    }
    // array and struct values
    if (typ.copy) {
        var clone = typ.zero();
        typ.copy(clone, v);
        return clone;
    }
    return v;
}

/* $isASCII reports whether string s contains only ASCII characters. */
var $isASCII = s => {
    for (var i = 0; i < s.length; i++) {
        if (s.charCodeAt(i) >= 128) {
            return false;
        }
    }
    return true;
};

$packages["github.com/gopherjs/gopherjs/js"] = (function() {
	var $pkg = {}, $init, Object, Error, sliceType, ptrType, ptrType$1, init;
	Object = $pkg.Object = $newType(0, $kindStruct, "js.Object", true, "github.com/gopherjs/gopherjs/js", true, function(object_) {
		this.$val = this;
		if (arguments.length === 0) {
			this.object = null;
			return;
		}
		this.object = object_;
	});
	Error = $pkg.Error = $newType(0, $kindStruct, "js.Error", true, "github.com/gopherjs/gopherjs/js", true, function(Object_) {
		this.$val = this;
		if (arguments.length === 0) {
			this.Object = null;
			return;
		}
		this.Object = Object_;
	});
	sliceType = $sliceType($emptyInterface);
	ptrType = $ptrType(Object);
	ptrType$1 = $ptrType(Error);
	Object.ptr.prototype.Get = function(key) {
		var key, o;
		o = this;
		return o.object[$externalize(key, $String)];
	};
	Object.prototype.Get = function(key) { return this.$val.Get(key); };
	Object.ptr.prototype.Set = function(key, value) {
		var key, o, value;
		o = this;
		o.object[$externalize(key, $String)] = $externalize(value, $emptyInterface);
	};
	Object.prototype.Set = function(key, value) { return this.$val.Set(key, value); };
	Object.ptr.prototype.Delete = function(key) {
		var key, o;
		o = this;
		delete o.object[$externalize(key, $String)];
	};
	Object.prototype.Delete = function(key) { return this.$val.Delete(key); };
	Object.ptr.prototype.Length = function() {
		var o;
		o = this;
		return $parseInt(o.object.length);
	};
	Object.prototype.Length = function() { return this.$val.Length(); };
	Object.ptr.prototype.Index = function(i) {
		var i, o;
		o = this;
		return o.object[i];
	};
	Object.prototype.Index = function(i) { return this.$val.Index(i); };
	Object.ptr.prototype.SetIndex = function(i, value) {
		var i, o, value;
		o = this;
		o.object[i] = $externalize(value, $emptyInterface);
	};
	Object.prototype.SetIndex = function(i, value) { return this.$val.SetIndex(i, value); };
	Object.ptr.prototype.Call = function(name, args) {
		var args, name, o, obj;
		o = this;
		return (obj = o.object, obj[$externalize(name, $String)].apply(obj, $externalize(args, sliceType)));
	};
	Object.prototype.Call = function(name, args) { return this.$val.Call(name, args); };
	Object.ptr.prototype.Invoke = function(args) {
		var args, o;
		o = this;
		return o.object.apply(undefined, $externalize(args, sliceType));
	};
	Object.prototype.Invoke = function(args) { return this.$val.Invoke(args); };
	Object.ptr.prototype.New = function(args) {
		var args, o;
		o = this;
		return new ($global.Function.prototype.bind.apply(o.object, [undefined].concat($externalize(args, sliceType))));
	};
	Object.prototype.New = function(args) { return this.$val.New(args); };
	Object.ptr.prototype.Bool = function() {
		var o;
		o = this;
		return !!(o.object);
	};
	Object.prototype.Bool = function() { return this.$val.Bool(); };
	Object.ptr.prototype.String = function() {
		var o;
		o = this;
		return $internalize(o.object, $String);
	};
	Object.prototype.String = function() { return this.$val.String(); };
	Object.ptr.prototype.Int = function() {
		var o;
		o = this;
		return $parseInt(o.object) >> 0;
	};
	Object.prototype.Int = function() { return this.$val.Int(); };
	Object.ptr.prototype.Int64 = function() {
		var o;
		o = this;
		return $internalize(o.object, $Int64);
	};
	Object.prototype.Int64 = function() { return this.$val.Int64(); };
	Object.ptr.prototype.Uint64 = function() {
		var o;
		o = this;
		return $internalize(o.object, $Uint64);
	};
	Object.prototype.Uint64 = function() { return this.$val.Uint64(); };
	Object.ptr.prototype.Float = function() {
		var o;
		o = this;
		return $parseFloat(o.object);
	};
	Object.prototype.Float = function() { return this.$val.Float(); };
	Object.ptr.prototype.Interface = function() {
		var o;
		o = this;
		return $internalize(o.object, $emptyInterface);
	};
	Object.prototype.Interface = function() { return this.$val.Interface(); };
	Object.ptr.prototype.Unsafe = function() {
		var o;
		o = this;
		return o.object;
	};
	Object.prototype.Unsafe = function() { return this.$val.Unsafe(); };
	Error.ptr.prototype.Error = function() {
		var err;
		err = this;
		return "JavaScript error: " + $internalize(err.Object.message, $String);
	};
	Error.prototype.Error = function() { return this.$val.Error(); };
	Error.ptr.prototype.Stack = function() {
		var err;
		err = this;
		return $internalize(err.Object.stack, $String);
	};
	Error.prototype.Stack = function() { return this.$val.Stack(); };
	init = function() {
		var e;
		e = new Error.ptr(null);
		$unused(e);
	};
	ptrType.methods = [{prop: "Get", name: "Get", pkg: "", typ: $funcType([$String], [ptrType], false)}, {prop: "Set", name: "Set", pkg: "", typ: $funcType([$String, $emptyInterface], [], false)}, {prop: "Delete", name: "Delete", pkg: "", typ: $funcType([$String], [], false)}, {prop: "Length", name: "Length", pkg: "", typ: $funcType([], [$Int], false)}, {prop: "Index", name: "Index", pkg: "", typ: $funcType([$Int], [ptrType], false)}, {prop: "SetIndex", name: "SetIndex", pkg: "", typ: $funcType([$Int, $emptyInterface], [], false)}, {prop: "Call", name: "Call", pkg: "", typ: $funcType([$String, sliceType], [ptrType], true)}, {prop: "Invoke", name: "Invoke", pkg: "", typ: $funcType([sliceType], [ptrType], true)}, {prop: "New", name: "New", pkg: "", typ: $funcType([sliceType], [ptrType], true)}, {prop: "Bool", name: "Bool", pkg: "", typ: $funcType([], [$Bool], false)}, {prop: "String", name: "String", pkg: "", typ: $funcType([], [$String], false)}, {prop: "Int", name: "Int", pkg: "", typ: $funcType([], [$Int], false)}, {prop: "Int64", name: "Int64", pkg: "", typ: $funcType([], [$Int64], false)}, {prop: "Uint64", name: "Uint64", pkg: "", typ: $funcType([], [$Uint64], false)}, {prop: "Float", name: "Float", pkg: "", typ: $funcType([], [$Float64], false)}, {prop: "Interface", name: "Interface", pkg: "", typ: $funcType([], [$emptyInterface], false)}, {prop: "Unsafe", name: "Unsafe", pkg: "", typ: $funcType([], [$Uintptr], false)}];
	ptrType$1.methods = [{prop: "Error", name: "Error", pkg: "", typ: $funcType([], [$String], false)}, {prop: "Stack", name: "Stack", pkg: "", typ: $funcType([], [$String], false)}];
	Object.init("github.com/gopherjs/gopherjs/js", [{prop: "object", name: "object", embedded: false, exported: false, typ: ptrType, tag: ""}]);
	Error.init("", [{prop: "Object", name: "Object", embedded: true, exported: true, typ: ptrType, tag: ""}]);
	$init = function() {
		$pkg.$init = function() {};
		/* */ var $f, $c = false, $s = 0, $r; if (this !== undefined && this.$blk !== undefined) { $f = this; $c = true; $s = $f.$s; $r = $f.$r; } s: while (true) { switch ($s) { case 0:
		init();
		/* */ } return; } if ($f === undefined) { $f = { $blk: $init }; } $f.$s = $s; $f.$r = $r; return $f;
	};
	$pkg.$init = $init;
	return $pkg;
})();
$packages["runtime"] = (function() {
	var $pkg = {}, $init, js, _type, TypeAssertionError, errorString, ptrType$1, ptrType$2, buildVersion, init, throw$1;
	js = $packages["github.com/gopherjs/gopherjs/js"];
	_type = $pkg._type = $newType(0, $kindStruct, "runtime._type", true, "runtime", false, function(str_) {
		this.$val = this;
		if (arguments.length === 0) {
			this.str = "";
			return;
		}
		this.str = str_;
	});
	TypeAssertionError = $pkg.TypeAssertionError = $newType(0, $kindStruct, "runtime.TypeAssertionError", true, "runtime", true, function(_interface_, concrete_, asserted_, missingMethod_) {
		this.$val = this;
		if (arguments.length === 0) {
			this._interface = ptrType$1.nil;
			this.concrete = ptrType$1.nil;
			this.asserted = ptrType$1.nil;
			this.missingMethod = "";
			return;
		}
		this._interface = _interface_;
		this.concrete = concrete_;
		this.asserted = asserted_;
		this.missingMethod = missingMethod_;
	});
	errorString = $pkg.errorString = $newType(8, $kindString, "runtime.errorString", true, "runtime", false, null);
	ptrType$1 = $ptrType(_type);
	ptrType$2 = $ptrType(TypeAssertionError);
	_type.ptr.prototype.string = function() {
		var t;
		t = this;
		return t.str;
	};
	_type.prototype.string = function() { return this.$val.string(); };
	_type.ptr.prototype.pkgpath = function() {
		var t;
		t = this;
		return "";
	};
	_type.prototype.pkgpath = function() { return this.$val.pkgpath(); };
	TypeAssertionError.ptr.prototype.RuntimeError = function() {
	};
	TypeAssertionError.prototype.RuntimeError = function() { return this.$val.RuntimeError(); };
	TypeAssertionError.ptr.prototype.Error = function() {
		var as, cs, e, inter, msg;
		e = this;
		inter = "interface";
		if (!(e._interface === ptrType$1.nil)) {
			inter = e._interface.string();
		}
		as = e.asserted.string();
		if (e.concrete === ptrType$1.nil) {
			return "interface conversion: " + inter + " is nil, not " + as;
		}
		cs = e.concrete.string();
		if (e.missingMethod === "") {
			msg = "interface conversion: " + inter + " is " + cs + ", not " + as;
			if (cs === as) {
				if (!(e.concrete.pkgpath() === e.asserted.pkgpath())) {
					msg = msg + (" (types from different packages)");
				} else {
					msg = msg + (" (types from different scopes)");
				}
			}
			return msg;
		}
		return "interface conversion: " + cs + " is not " + as + ": missing method " + e.missingMethod;
	};
	TypeAssertionError.prototype.Error = function() { return this.$val.Error(); };
	init = function() {
		var e, jsPkg;
		jsPkg = $packages[$externalize("github.com/gopherjs/gopherjs/js", $String)];
		$jsObjectPtr = jsPkg.Object.ptr;
		$jsErrorPtr = jsPkg.Error.ptr;
		$throwRuntimeError = throw$1;
		buildVersion = $internalize($goVersion, $String);
		e = $ifaceNil;
		e = new TypeAssertionError.ptr(ptrType$1.nil, ptrType$1.nil, ptrType$1.nil, "");
		$unused(e);
	};
	errorString.prototype.RuntimeError = function() {
		var e;
		e = this.$val;
	};
	$ptrType(errorString).prototype.RuntimeError = function() { return new errorString(this.$get()).RuntimeError(); };
	errorString.prototype.Error = function() {
		var e;
		e = this.$val;
		return "runtime error: " + (e);
	};
	$ptrType(errorString).prototype.Error = function() { return new errorString(this.$get()).Error(); };
	throw$1 = function(s) {
		var s;
		$panic(new errorString((s)));
	};
	ptrType$1.methods = [{prop: "string", name: "string", pkg: "runtime", typ: $funcType([], [$String], false)}, {prop: "pkgpath", name: "pkgpath", pkg: "runtime", typ: $funcType([], [$String], false)}];
	ptrType$2.methods = [{prop: "RuntimeError", name: "RuntimeError", pkg: "", typ: $funcType([], [], false)}, {prop: "Error", name: "Error", pkg: "", typ: $funcType([], [$String], false)}];
	errorString.methods = [{prop: "RuntimeError", name: "RuntimeError", pkg: "", typ: $funcType([], [], false)}, {prop: "Error", name: "Error", pkg: "", typ: $funcType([], [$String], false)}];
	_type.init("runtime", [{prop: "str", name: "str", embedded: false, exported: false, typ: $String, tag: ""}]);
	TypeAssertionError.init("runtime", [{prop: "_interface", name: "_interface", embedded: false, exported: false, typ: ptrType$1, tag: ""}, {prop: "concrete", name: "concrete", embedded: false, exported: false, typ: ptrType$1, tag: ""}, {prop: "asserted", name: "asserted", embedded: false, exported: false, typ: ptrType$1, tag: ""}, {prop: "missingMethod", name: "missingMethod", embedded: false, exported: false, typ: $String, tag: ""}]);
	$init = function() {
		$pkg.$init = function() {};
		/* */ var $f, $c = false, $s = 0, $r; if (this !== undefined && this.$blk !== undefined) { $f = this; $c = true; $s = $f.$s; $r = $f.$r; } s: while (true) { switch ($s) { case 0:
		$r = js.$init(); /* */ $s = 1; case 1: if($c) { $c = false; $r = $r.$blk(); } if ($r && $r.$blk !== undefined) { break s; }
		buildVersion = "";
		init();
		/* */ } return; } if ($f === undefined) { $f = { $blk: $init }; } $f.$s = $s; $f.$r = $r; return $f;
	};
	$pkg.$init = $init;
	return $pkg;
})();
$packages["main"] = (function() {
	var $pkg = {}, $init, js, funcType, main;
	js = $packages["github.com/gopherjs/gopherjs/js"];
	funcType = $funcType([], [], false);
	main = function() {
		var button, document;
		document = $global.document;
		button = document.createElement($externalize("button", $String));
		button.innerHTML = $externalize("Click me", $String);
		document.body.appendChild(button);
		button.addEventListener($externalize("click", $String), $externalize((function() {
			var p;
			p = document.createElement($externalize("p", $String));
			p.innerHTML = $externalize("Hello, big world!", $String);
			document.body.appendChild(p);
		}), funcType));
	};
	$init = function() {
		$pkg.$init = function() {};
		/* */ var $f, $c = false, $s = 0, $r; if (this !== undefined && this.$blk !== undefined) { $f = this; $c = true; $s = $f.$s; $r = $f.$r; } s: while (true) { switch ($s) { case 0:
		$r = js.$init(); /* */ $s = 1; case 1: if($c) { $c = false; $r = $r.$blk(); } if ($r && $r.$blk !== undefined) { break s; }
		if ($pkg === $mainPkg) {
			main();
			$mainFinished = true;
		}
		/* */ } return; } if ($f === undefined) { $f = { $blk: $init }; } $f.$s = $s; $f.$r = $r; return $f;
	};
	$pkg.$init = $init;
	return $pkg;
})();
$synthesizeMethods();
$initAllLinknames();
var $mainPkg = $packages["main"];
$packages["runtime"].$init();
$go($mainPkg.$init, []);
$flushConsole();

}).call(this);
//# sourceMappingURL=hello.js.map
