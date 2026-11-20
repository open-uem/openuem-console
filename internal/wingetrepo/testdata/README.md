# Test data

- `test.msi`: an MSI package built with `msibuild` (msitools) with a product name,
  manufacturer, version and product code.
- `test.exe`: a minimal Windows executable with the version resource of `test-exe.rc`,
  built with LLVM:

  ```sh
  printf 'int mainCRTStartup(void) { return 0; }\n' > stub.c
  clang --target=x86_64-pc-windows-msvc -c -O2 -fno-stack-protector stub.c -o stub.obj
  llvm-rc -no-preprocess -fo test-exe.res test-exe.rc
  lld-link /nologo /entry:mainCRTStartup /subsystem:console /nodefaultlib /out:test.exe stub.obj test-exe.res
  ```

- `test-inno.exe`, `test-nsis.exe`: `test.exe` followed by the signatures Inno Setup and
  NSIS installers carry after the executable (`Inno Setup Setup Data (6.2.2)` and
  `\xEF\xBE\xAD\xDENullsoftInst`).
- `test-burn.exe`: `test.exe` with a `.wixburn` section, like WiX Burn bundles:
  `llvm-objcopy --add-section .wixburn=data.bin test.exe test-burn.exe`.
