### Usefull Commands and Instructions

- To check using the command line if a server is running on a certain port:
```bash
nc -zv localhost 80
```

- To send a packet of information to the server using the netcat command use:
```bash
echo -n “Ground Control For Major Tom” | nc localhost 8080
```

### Basic terminology

- `byte` is the basic unit of data used to store and represent information. A byte is a group of 8 bits, which can represent 256 different values (from 0 to 255). When dealing with network communication, data is typically read and written as raw bytes, regardless of the type of data it represents (text, numbers, etc.).

-  `bit` is the most basic unit of data in computing and can hold one of two possible values: `0` or `1`. 8 `bits` you can represent 256 (2^8) unique numbers which correspond to characters in the `ASCII` character set. Ex: the byte `01001000` represents `H` in `ASCII`. 



### Converting Number to Binary
- Example convert 72 to binary

1. Binary place values are:
    ```
    128 64 32 16 8 4 2 1
    ```

2. Start with the highest place value `128`. Since 72 is less than 128, put a `0`.
3. Compare 72 with `64`. Since 72 is greater than 64, put a `1` and and subtract 64 from 72, leaving 8.
4. Continue until the end.

Therefore `72` in binary is `01001000`

### ASCII Encoding

### Converting from byte array to string

- Suppose we have a byte array like the following:
    ```go
    [72, 101, 108, 108, 111]
    ```

- Do a look up in the `ASCII` table to determine which character the number coresponds to. 

