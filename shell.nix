{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go_1_27
    libwebp   # cwebp
  ];

  shellHook = ''
    export PATH="$HOME/.local/bin:$PATH"
  '';
}
