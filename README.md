# Redux
A simple home cloud with the goal to be both Simple and Secure.

## 1. Design Goals
### 1.1 Simplicity
This cloud will implement a minimal set of features with straightforward solutions. 
One of the main goals in the implementation is to avoid feature bloat and overengineering
and eventually reach a 'done' state in which the application will only recieve maintenance updates.

### 1.2 Security
Security will be taken very seriously in this project, as it will handle potentially sensitive information.

#### 1.2.1 Authorization & Authentication
Authorization will be done by expanding a PBKDF-2 Hash of the password on the client side, which will then be combined with a salt that is unique for every user and hashed again using SHA512 on the server. 
This guarantees that the plaintext password is never transferred over the network and the authorization secret is not present in the database.
Once the authorization has been successfully completed, a RSA4096 Signed Json Web Token (JWT) will be issued to the user that completed the authorization. The JWT may then be used in further requests to authenticate with the server.

#### 1.2.2 File Encryption
Files will be Encrypted/Decrypted when a file is Uploaded/Downloaded, on the client. The files will be encrypted using AES256, which will use a PBKDF-2 Hash of the users password as the encryption key. This ensures that the files on the server will never be unencrypted, and even when the server is compromised, the files are not.

#### 1.2.3 Changeroot
After opening a handle to the sqlite database file and reading the X509-Keypair, the application will use the 
changeroot syscall to jail itself to a 'virtual filesystem' located under fs-root/files/, where Fs-Root is the value of the --fs-root CLI variable. This mitigates the risk of users reading and writing parts of the filesystem, that they should not be interacting with.

#### 1.2.4 Priviledge Dropping
The app will drop root priviledges after entering the changeroot environment and use the setresuid syscall to switch to a user that has as little permissions as possible.

#### 1.2.5 Docker
The entire app will run inside of a docker container, which has the user with the minimal permissions already preconfigured. The container also makes the app more portable and provides an additional layer of sandboxing