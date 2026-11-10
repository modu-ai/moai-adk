//! Color profile and SGR sequences, matching what the Go statusline emits through
//! lipgloss/termenv v0.16.0 and go-colorful v1.4.1 (#2350). The arithmetic is kept in the
//! same order as those libraries so the bytes match the recorded goldens: a truecolor
//! channel is `uint8(v * (1/255) * 255)` (which turns 244 into 243), and the ANSI256 path
//! is termenv's `hexToANSI256Color`, HSLuv distance included.

// The HSLuv constants are go-colorful's literals copied digit for digit, so they can be
// compared against the source; the extra digits round to the same f64.
#![allow(clippy::excessive_precision)]

/// Which palette the statusline writes. Claude Code reads stdout through a pipe, so the
/// profile is chosen from `COLORTERM` alone, never from TTY detection.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Profile {
    TrueColor,
    Ansi256,
}

impl Profile {
    pub fn from_colorterm(colorterm: &str) -> Profile {
        match colorterm.trim().to_ascii_lowercase().as_str() {
            "truecolor" | "24bit" => Profile::TrueColor,
            _ => Profile::Ansi256,
        }
    }

    /// The foreground SGR parameter for a `#RRGGBB` color, without `ESC[` and `m`.
    pub fn foreground(self, hex: &str) -> String {
        let (r, g, b) = hex_rgb(hex);
        match self {
            Profile::TrueColor => format!(
                "38;2;{};{};{}",
                channel_255(r),
                channel_255(g),
                channel_255(b)
            ),
            Profile::Ansi256 => format!("38;5;{}", ansi256(r, g, b)),
        }
    }

    /// `ESC[<fg>m` + text + `ESC[0m`, the shape a lipgloss foreground style renders.
    pub fn paint(self, hex: &str, text: &str) -> String {
        format!("\x1b[{}m{text}\x1b[0m", self.foreground(hex))
    }
}

/// `#RRGGBB` → integer channels; anything else reads as black (go-colorful/gradient.go).
pub fn hex_rgb(hex: &str) -> (u8, u8, u8) {
    let digits = hex.strip_prefix('#').unwrap_or(hex);
    match u32::from_str_radix(digits, 16) {
        Ok(v) if digits.len() == 6 => ((v >> 16) as u8, (v >> 8) as u8, v as u8),
        _ => (0, 0, 0),
    }
}

/// go-colorful parses a channel as `v * (1/255)`; termenv writes it back as `uint8(f*255)`.
fn unit(v: u8) -> f64 {
    f64::from(v) * (1.0 / 255.0)
}

fn channel_255(v: u8) -> u8 {
    (unit(v) * 255.0) as u8
}

/// termenv `hexToANSI256Color`, including its gray-index computation from the cube
/// indices (which always lands on gray 232; kept as is for byte equality).
fn ansi256(r: u8, g: u8, b: u8) -> u16 {
    // The goldens come from an arm64 Go build, where the compiler fuses `c*255 - 35` into
    // one FNMSUBD (ARM64.rules `(FSUBD (FMULD x y) a)`): the subtraction sees the unrounded
    // product, so channel 155 lands on cube index 2, not 3. The comparisons still read the
    // rounded product.
    let v2ci = |c: f64| -> i32 {
        let v = c * 255.0;
        if v < 48.0 {
            0
        } else if v < 115.0 {
            1
        } else {
            (c.mul_add(255.0, -35.0) / 40.0) as i32
        }
    };
    let (rf, gf, bf) = (unit(r), unit(g), unit(b));
    let (ri, gi, bi) = (v2ci(rf), v2ci(gf), v2ci(bf));
    let ci = 36 * ri + 6 * gi + bi;
    const I2CV: [i32; 6] = [0, 0x5f, 0x87, 0xaf, 0xd7, 0xff];
    let (cr, cg, cb) = (I2CV[ri as usize], I2CV[gi as usize], I2CV[bi as usize]);
    let average = (ri + gi + bi) / 3;
    let gray_idx = if average > 238 {
        23
    } else {
        (average - 3) / 10
    };
    let gv = 8 + 10 * gray_idx;
    let here = [rf, gf, bf];
    let cube = [cr, cg, cb].map(|c| f64::from(c) / 255.0);
    let gray = [f64::from(gv) / 255.0; 3];
    if distance_hsluv(here, cube) <= distance_hsluv(here, gray) {
        (16 + ci) as u16
    } else {
        (232 + gray_idx) as u16
    }
}

// ---- go-colorful HSLuv (hsluv.go, colors.go) ----

const HSLUV_D65: [f64; 3] = [0.95045592705167, 1.0, 1.089057750759878];
const M: [[f64; 3]; 3] = [
    [
        3.2409699419045214,
        -1.5373831775700935,
        -0.49861076029300328,
    ],
    [
        -0.96924363628087983,
        1.8759675015077207,
        0.041555057407175613,
    ],
    [
        0.055630079696993609,
        -0.20397695888897657,
        1.0569715142428786,
    ],
];
const KAPPA: f64 = 903.2962962962963;
const EPSILON: f64 = 0.0088564516790356308;

fn distance_hsluv(a: [f64; 3], b: [f64; 3]) -> f64 {
    let (h1, s1, l1) = hsluv(a);
    let (h2, s2, l2) = hsluv(b);
    (sq((h1 - h2) / 100.0) + sq(s1 - s2) + sq(l1 - l2)).sqrt()
}

fn sq(v: f64) -> f64 {
    v * v
}

fn clamp01(v: f64) -> f64 {
    v.clamp(0.0, 1.0)
}

fn linearize(v: f64) -> f64 {
    if v <= 0.04045 {
        v / 12.92
    } else {
        ((v + 0.055) / 1.055).powf(2.4)
    }
}

fn xyz_to_uv(x: f64, y: f64, z: f64) -> (f64, f64) {
    let denom = x + 15.0 * y + 3.0 * z;
    if denom == 0.0 {
        (0.0, 0.0)
    } else {
        (4.0 * x / denom, 9.0 * y / denom)
    }
}

fn hsluv([r, g, b]: [f64; 3]) -> (f64, f64, f64) {
    let (r, g, b) = (linearize(r), linearize(g), linearize(b));
    let x = 0.41239079926595948 * r + 0.35758433938387796 * g + 0.18048078840183429 * b;
    let y = 0.21263900587151036 * r + 0.71516867876775593 * g + 0.072192315360733715 * b;
    let z = 0.019330818715591851 * r + 0.11919477979462599 * g + 0.95053215224966058 * b;
    let w = HSLUV_D65;
    let l = if y / w[1] <= 6.0 / 29.0 * 6.0 / 29.0 * 6.0 / 29.0 {
        y / w[1] * (29.0 / 3.0 * 29.0 / 3.0 * 29.0 / 3.0) / 100.0
    } else {
        1.16 * (y / w[1]).cbrt() - 0.16
    };
    let (ubis, vbis) = xyz_to_uv(x, y, z);
    let (un, vn) = xyz_to_uv(w[0], w[1], w[2]);
    let u = 13.0 * l * (ubis - un);
    let v = 13.0 * l * (vbis - vn);
    // LuvToLuvLCh
    let h = if (v - u).abs() > 1e-4 && u.abs() > 1e-4 {
        (57.29577951308232087721 * v.atan2(u) + 360.0) % 360.0
    } else {
        0.0
    };
    let c = (sq(u) + sq(v)).sqrt();
    // LuvLChToHSLuv
    let (c, l) = (c * 100.0, l * 100.0);
    let s = if !(0.00000001..=99.9999999).contains(&l) {
        0.0
    } else {
        c / max_chroma_for_lh(l, h) * 100.0
    };
    (h, clamp01(s / 100.0), clamp01(l / 100.0))
}

fn max_chroma_for_lh(l: f64, h: f64) -> f64 {
    let h_rad = h / 360.0 * std::f64::consts::PI * 2.0;
    let mut min_length = f64::MAX;
    for [x, y] in bounds(l) {
        let length = y / (h_rad.sin() - x * h_rad.cos());
        if length > 0.0 && length < min_length {
            min_length = length;
        }
    }
    min_length
}

fn bounds(l: f64) -> [[f64; 2]; 6] {
    let sub1 = (l + 16.0).powf(3.0) / 1560896.0;
    let sub2 = if sub1 > EPSILON { sub1 } else { l / KAPPA };
    let mut ret = [[0.0; 2]; 6];
    for (i, m) in M.iter().enumerate() {
        for k in 0..2 {
            let kf = f64::from(k as u8);
            let top1 = (284517.0 * m[0] - 94839.0 * m[2]) * sub2;
            let top2 = (838422.0 * m[2] + 769860.0 * m[1] + 731718.0 * m[0]) * l * sub2
                - 769860.0 * kf * l;
            let bottom = (632260.0 * m[2] - 126452.0 * m[1]) * sub2 + 126452.0 * kf;
            ret[i * 2 + k] = [top1 / bottom, top2 / bottom];
        }
    }
    ret
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn truecolor_channel_keeps_the_float_truncation() {
        // Mocha text #CDD6F4: termenv writes 205;214;243, not 244 (golden a01.ansi).
        assert_eq!(Profile::TrueColor.foreground("#CDD6F4"), "38;2;205;214;243");
        assert_eq!(Profile::TrueColor.foreground("#A6E3A1"), "38;2;166;227;161");
    }

    #[test]
    fn colorterm_selects_the_profile() {
        assert_eq!(Profile::from_colorterm(" TrueColor "), Profile::TrueColor);
        assert_eq!(Profile::from_colorterm("24bit"), Profile::TrueColor);
        assert_eq!(Profile::from_colorterm(""), Profile::Ansi256);
        assert_eq!(Profile::from_colorterm("yes"), Profile::Ansi256);
    }

    #[test]
    fn ansi256_cube_index_follows_the_fused_go_subtraction() {
        // Channel 155 is the 2/3 boundary: fused, (155*(1/255))*255 - 35 is just below 120,
        // so Go picks index 2 (golden b01.ansi256: 222). 156 is past it either way.
        assert_eq!(ansi256(250, 203, 155), 222);
        assert_eq!(ansi256(250, 203, 156), 223);
    }

    #[test]
    fn malformed_hex_reads_as_black() {
        assert_eq!(hex_rgb("#12"), (0, 0, 0));
        assert_eq!(hex_rgb("#GGGGGG"), (0, 0, 0));
    }
}
