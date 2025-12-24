data "utilities_bcrypt_hash" "example" {
  plaintext = "my-secret-password"
  cost      = 10
}

output "utilities_bcrypt_hash" {
  value = {
    id        = data.utilities_bcrypt_hash.example.id
    plaintext = data.utilities_bcrypt_hash.example.plaintext
    cost      = data.utilities_bcrypt_hash.example.cost
    hash      = data.utilities_bcrypt_hash.example.hash
  }
}

