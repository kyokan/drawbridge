pragma solidity 0.4.24;

library BytesBuffer {
    struct Buffer {
        bytes data;
        uint ptr;
    }

    function putByte(Buffer self, bytes1 b) internal {
        require(self.data.length >= self.ptr + 1);
        self.data[self.ptr] = b;
        self.ptr++;
    }

    function putUint256Le(Buffer self, uint n) internal {
        require(self.data.length >= self.ptr + 32);

        bytes memory b = new bytes(32);
        assembly {
            mstore(add(b, 32), n)
        }

        for (uint i = 0; i < b.length; i++) {
            self.data[self.ptr + i] = b[i];
        }

        concat(self, b);
    }

    function putAddress(Buffer self, address a) internal {
        require(self.data.length >= self.ptr + 20);

        bytes memory b;

        assembly {
            let m := mload(0x40)
            mstore(add(m, 20), xor(0x140000000000000000000000000000000000000000, a))
            mstore(0x40, add(m, 52))
            b := m
        }

        concat(self, b);
    }

    function putBytes32(Buffer self, bytes32 b) internal {
        require(self.data.length >= self.ptr + 32);

        for (uint i = 0; i < 32; i++) {
            self.data[self.ptr + i] = b[i];
        }

        self.ptr += 32;
    }
    
    function trimmed(Buffer self) internal returns (bytes) {
        if (self.data.length == self.ptr) {
            return self.data;
        }
        
        bytes memory out = new bytes(self.ptr);
        for (uint i = 0; i < self.ptr; i++) {
            out[i] = self.data[i];
        }
        
        return out;
    }

    function putBytes(Buffer self, bytes b) internal {
        require(self.data.length >= self.ptr + b.length);

        concat(self, b);
    }

    function concat(Buffer self, bytes b) private {
        for (uint i = 0; i < b.length; i++) {
            self.data[self.ptr + i] = b[i];
        }

        self.ptr += b.length;
    }
}