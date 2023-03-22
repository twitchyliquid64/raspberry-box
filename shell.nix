let
  # Pinned nixpkgs, deterministic
  nixpkgs = import (fetchTarball ("https://github.com/NixOS/nixpkgs/archive/refs/heads/nixos-21.05.tar.gz")) { };

  # Rolling updates, not deterministic.
  # pkgs = import (fetchTarball("channel:nixpkgs-unstable")) {};
in
nixpkgs.mkShell {
  buildInputs = [
    nixpkgs.pkgconfig
    nixpkgs.sqlite
    nixpkgs.glibc
    nixpkgs.go
    nixpkgs.gpgme
    nixpkgs.gnupg
    nixpkgs.btrfs-progs
    nixpkgs.lvm2
  ];

  CGO_LDFLAGS="-L${nixpkgs.stdenv.cc.cc.lib}/lib -Xlinker -rpath=${nixpkgs.glibc}/lib";
}
