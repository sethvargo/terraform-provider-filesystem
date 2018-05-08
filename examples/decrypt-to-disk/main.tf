// Reference an encrypted file from a Google storage bucket
data "google_storage_object_signed_url" "encrypted-file" {
  bucket = "my-storage-bucket"
  path   = "encrypted-file.txt"
}

// Download the encrypted file
data "http" "download-file" {
  url = "${data.google_storage_object_signed_url.encrypted-file.signed_url}"
}

// Decrypted the value
data "google_kms_secret" "decrypt-file" {
  crypto_key = "my-crypto-key"
  ciphertext = "${data.http.download-file.body}"
}

// Write the decrypted value to disk
resource "filesystem_file_writer" "save-file" {
  path     = "decrypted-file.txt"
  contents = "${data.google_kms_secret.decrypt-file.plaintext}"
  mode     = "0600"
}
