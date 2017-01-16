Crypto Chat 
===========
This is a Go learning project that tries to implement a secure chat with a centralized server and E2E encryption of messages. 
The server utilises elliptic curve cryptography for TLS 1.2 and JWT tokens, and the client uses it for message encryption and decryption.

Content
-------
* Server used for client discovery and storing of encrypted messages 
* Client that is able to encrypt and decrypt sent and received messages, and verify their authenticity

Libraries
---------
* [NaCL](https://godoc.org/golang.org/x/crypto/nacl/box)
* [Go-SQLite3](https://github.com/mattn/go-sqlite3)
* [JWT Go](https://github.com/dgrijalva/jwt-go)
* [Context](https://github.com/gorilla/context)

License
-------

    This is free and unencumbered software released into the public domain.
    
    Anyone is free to copy, modify, publish, use, compile, sell, or
    distribute this software, either in source code form or as a compiled
    binary, for any purpose, commercial or non-commercial, and by any
    means.
    
    In jurisdictions that recognize copyright laws, the author or authors
    of this software dedicate any and all copyright interest in the
    software to the public domain. We make this dedication for the benefit
    of the public at large and to the detriment of our heirs and
    successors. We intend this dedication to be an overt act of
    relinquishment in perpetuity of all present and future rights to this
    software under copyright law.
    
    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
    EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
    MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
    IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
    OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
    ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
    OTHER DEALINGS IN THE SOFTWARE.
    
    For more information, please refer to <http://unlicense.org>