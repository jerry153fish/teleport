// Copyright 2022 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

use std::{env, path::PathBuf};

use cbindgen::{Builder, Language};

fn main() {
    let crate_dir: PathBuf = env::var_os("CARGO_MANIFEST_DIR").unwrap().into();

    let bindings = Builder::new()
        .with_language(Language::C)
        .with_crate(&crate_dir)
        .generate()
        .unwrap();

    let out = tempfile::NamedTempFile::new_in(&crate_dir).unwrap();
    bindings.write(&out);
    out.persist(crate_dir.join("librdprs.h")).unwrap();
}
