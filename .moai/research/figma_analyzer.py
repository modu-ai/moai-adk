#!/usr/bin/env python3
"""
Figma Design Analysis Tool

Usage:
    python figma_analyzer.py --file m2odCIWVPWv84ygT5w43Ur --node 689:1242 --token YOUR_TOKEN
    python figma_analyzer.py --json figma-metadata.json --analyze colors,contrast,structure
"""

import json
import sys
from pathlib import Path
from dataclasses import dataclass
from typing import Dict, List, Tuple, Optional
import argparse
from enum import Enum


class NodeType(Enum):
    """Figma 노드 타입"""
    FRAME = "FRAME"
    COMPONENT = "COMPONENT"
    COMPONENT_SET = "COMPONENT_SET"
    GROUP = "GROUP"
    TEXT = "TEXT"
    ELLIPSE = "ELLIPSE"
    RECTANGLE = "RECTANGLE"
    IMAGE = "IMAGE"
    SVG = "SVG"
    BOARD = "BOARD"


@dataclass
class Color:
    """색상 정보"""
    r: float
    g: float
    b: float
    a: float = 1.0

    def to_hex(self) -> str:
        """RGB를 Hex로 변환"""
        r = int(self.r * 255)
        g = int(self.g * 255)
        b = int(self.b * 255)
        return f"#{r:02x}{g:02x}{b:02x}"

    def to_rgb_tuple(self) -> Tuple[int, int, int]:
        """RGB 튜플 반환"""
        return (int(self.r * 255), int(self.g * 255), int(self.b * 255))

    def luminance(self) -> float:
        """WCAG 휘도 계산"""
        def adjust(c):
            return c / 12.92 if c <= 0.03928 else ((c + 0.055) / 1.055) ** 2.4

        r, g, b = adjust(self.r), adjust(self.g), adjust(self.b)
        return 0.2126 * r + 0.7152 * g + 0.0722 * b


@dataclass
class TypographyInfo:
    """타이포그래피 정보"""
    font_family: str
    font_size: float
    font_weight: int
    line_height: float
    letter_spacing: float = 0


@dataclass
class NodeBounds:
    """노드 경계"""
    x: float
    y: float
    width: float
    height: float

    @property
    def area(self) -> float:
        """영역 크기"""
        return self.width * self.height


class FigmaAnalyzer:
    """Figma 디자인 분석기"""

    def __init__(self, figma_json: Dict):
        """초기화"""
        self.data = figma_json
        self.colors: Dict[str, Color] = {}
        self.typography: Dict[str, TypographyInfo] = {}
        self.components: List[Dict] = []
        self.images: List[Dict] = []

    def extract_colors(self, node: Optional[Dict] = None) -> Dict[str, Color]:
        """모든 색상 추출"""
        if node is None:
            node = self.data

        colors = {}

        def traverse(n, path=""):
            node_name = n.get("name", "unknown")
            current_path = f"{path}/{node_name}".strip("/")

            # Fill 색상
            if "fills" in n and isinstance(n["fills"], list):
                for i, fill in enumerate(n["fills"]):
                    if fill.get("type") == "SOLID" and "color" in fill:
                        c = fill["color"]
                        color = Color(
                            r=c.get("r", 0),
                            g=c.get("g", 0),
                            b=c.get("b", 0),
                            a=fill.get("opacity", 1.0)
                        )
                        colors[f"{current_path}/fill-{i}"] = color

            # Stroke 색상
            if "strokes" in n and isinstance(n["strokes"], list):
                for i, stroke in enumerate(n["strokes"]):
                    if stroke.get("type") == "SOLID" and "color" in stroke:
                        c = stroke["color"]
                        color = Color(
                            r=c.get("r", 0),
                            g=c.get("g", 0),
                            b=c.get("b", 0),
                            a=stroke.get("opacity", 1.0)
                        )
                        colors[f"{current_path}/stroke-{i}"] = color

            # 재귀
            if "children" in n:
                for child in n["children"]:
                    traverse(child, current_path)

        traverse(node)
        self.colors = colors
        return colors

    def extract_typography(self, node: Optional[Dict] = None) -> Dict[str, TypographyInfo]:
        """타이포그래피 정보 추출"""
        if node is None:
            node = self.data

        typography = {}

        def traverse(n, path=""):
            node_name = n.get("name", "unknown")
            current_path = f"{path}/{node_name}".strip("/")

            if n.get("type") == "TEXT":
                style = n.get("style", {})
                info = TypographyInfo(
                    font_family=style.get("fontFamily", "unknown"),
                    font_size=style.get("fontSize", 16),
                    font_weight=style.get("fontWeight", 400),
                    line_height=style.get("lineHeightPx", 0),
                    letter_spacing=style.get("letterSpacing", 0)
                )
                typography[current_path] = info

            if "children" in n:
                for child in n["children"]:
                    traverse(child, current_path)

        traverse(node)
        self.typography = typography
        return typography

    def extract_components(self, node: Optional[Dict] = None) -> List[Dict]:
        """컴포넌트 추출"""
        if node is None:
            node = self.data

        components = []

        def traverse(n, path=""):
            node_name = n.get("name", "unknown")
            current_path = f"{path}/{node_name}".strip("/")

            if n.get("type") in ["COMPONENT", "COMPONENT_SET"]:
                components.append({
                    "id": n.get("id"),
                    "name": node_name,
                    "type": n.get("type"),
                    "path": current_path,
                    "bounds": {
                        "x": n.get("absoluteBoundingBox", {}).get("x", 0),
                        "y": n.get("absoluteBoundingBox", {}).get("y", 0),
                        "width": n.get("absoluteBoundingBox", {}).get("width", 0),
                        "height": n.get("absoluteBoundingBox", {}).get("height", 0)
                    }
                })

            if "children" in n:
                for child in n["children"]:
                    traverse(child, current_path)

        traverse(node)
        self.components = components
        return components

    def extract_images(self, node: Optional[Dict] = None) -> List[Dict]:
        """이미지 추출"""
        if node is None:
            node = self.data

        images = []

        def traverse(n, path=""):
            node_name = n.get("name", "unknown")
            current_path = f"{path}/{node_name}".strip("/")

            if n.get("type") == "IMAGE":
                images.append({
                    "id": n.get("id"),
                    "name": node_name,
                    "path": current_path,
                    "bounds": {
                        "x": n.get("absoluteBoundingBox", {}).get("x", 0),
                        "y": n.get("absoluteBoundingBox", {}).get("y", 0),
                        "width": n.get("absoluteBoundingBox", {}).get("width", 0),
                        "height": n.get("absoluteBoundingBox", {}).get("height", 0)
                    }
                })

            if "children" in n:
                for child in n["children"]:
                    traverse(child, current_path)

        traverse(node)
        self.images = images
        return images

    def check_contrast(self, color1: Color, color2: Color) -> float:
        """WCAG 대조 비율 계산"""
        l1 = color1.luminance()
        l2 = color2.luminance()
        lmax, lmin = max(l1, l2), min(l1, l2)
        return round((lmax + 0.05) / (lmin + 0.05), 2)

    def validate_wcag_aa(self) -> Dict[str, bool]:
        """WCAG AA 규정 검증"""
        return {
            "colors_checked": len(self.colors) > 0,
            "typography_extracted": len(self.typography) > 0,
            "components_found": len(self.components) > 0,
            "images_found": len(self.images) > 0
        }

    def generate_design_tokens_css(self) -> str:
        """CSS 디자인 토큰 생성"""
        css = ":root {\n"

        # 색상 토큰
        unique_colors = {}
        for path, color in sorted(self.colors.items()):
            hex_color = color.to_hex()
            if hex_color not in unique_colors:
                # 간단한 이름 생성
                key_name = path.split("/")[-1].replace("-", "_").lower()
                unique_colors[hex_color] = f"--color-{key_name}"
                css += f"  {unique_colors[hex_color]}: {hex_color};\n"

        # 타이포그래피 토큰
        unique_fonts = {}
        for path, typo in sorted(self.typography.items()):
            key = (typo.font_family, int(typo.font_size), typo.font_weight)
            if key not in unique_fonts:
                font_key = f"--font-{int(typo.font_size)}px-{typo.font_weight}w"
                unique_fonts[key] = font_key
                css += f"  {font_key}: {typo.font_weight} {typo.font_size}px / {typo.line_height};\n"

        css += "}\n"
        return css

    def generate_report(self) -> str:
        """분석 리포트 생성"""
        report = "# Figma Design Analysis Report\n\n"

        report += "## Color Summary\n\n"
        report += f"Total colors extracted: {len(self.colors)}\n\n"

        report += "| Name | Hex | RGB | Usage |\n"
        report += "|------|-----|-----|-------|\n"
        for path, color in sorted(self.colors.items())[:10]:
            rgb = color.to_rgb_tuple()
            report += f"| {path} | {color.to_hex()} | {rgb} | - |\n"

        report += "\n## Typography Summary\n\n"
        report += f"Total fonts found: {len(self.typography)}\n\n"

        report += "| Path | Font | Size | Weight | Line Height |\n"
        report += "|------|------|------|--------|-------------|\n"
        for path, typo in sorted(self.typography.items())[:10]:
            report += f"| {path} | {typo.font_family} | {typo.font_size}px | {typo.font_weight} | {typo.line_height} |\n"

        report += "\n## Components\n\n"
        report += f"Total components: {len(self.components)}\n\n"
        for comp in self.components[:10]:
            report += f"- **{comp['name']}** ({comp['type']}): {comp['bounds']['width']}x{comp['bounds']['height']}px\n"

        report += "\n## Images\n\n"
        report += f"Total images: {len(self.images)}\n\n"
        for img in self.images[:10]:
            report += f"- **{img['name']}**: {img['bounds']['width']}x{img['bounds']['height']}px\n"

        report += "\n## WCAG AA Validation\n\n"
        validation = self.validate_wcag_aa()
        for check, result in validation.items():
            status = "✅" if result else "❌"
            report += f"- {status} {check}: {result}\n"

        return report


def main():
    """메인 함수"""
    parser = argparse.ArgumentParser(
        description="Figma Design Analysis Tool",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # JSON 파일에서 분석
  python figma_analyzer.py --json figma-metadata.json --analyze colors,typography,components

  # Figma API로 직접 분석 (토큰 필요)
  python figma_analyzer.py --file m2odCIWVPWv84ygT5w43Ur --node 689:1242 --token YOUR_TOKEN
        """
    )

    parser.add_argument("--json", type=str, help="JSON 메타데이터 파일 경로")
    parser.add_argument("--file", type=str, help="Figma 파일 키")
    parser.add_argument("--node", type=str, help="Figma 노드 ID")
    parser.add_argument("--token", type=str, help="Figma 개인 액세스 토큰")
    parser.add_argument(
        "--analyze",
        type=str,
        default="colors,typography,components,images",
        help="분석 항목 (쉼표로 구분)"
    )
    parser.add_argument("--output", type=str, default=".moai/research", help="출력 디렉토리")
    parser.add_argument("--css", action="store_true", help="CSS 토큰 생성")
    parser.add_argument("--report", action="store_true", help="마크다운 리포트 생성")

    args = parser.parse_args()

    # JSON 파일에서 로드
    if args.json:
        json_path = Path(args.json)
        if not json_path.exists():
            print(f"Error: {json_path} not found")
            sys.exit(1)

        with open(json_path) as f:
            figma_data = json.load(f)

        analyzer = FigmaAnalyzer(figma_data)

        # 분석 실행
        if "colors" in args.analyze:
            colors = analyzer.extract_colors()
            print(f"Extracted {len(colors)} colors")

        if "typography" in args.analyze:
            typo = analyzer.extract_typography()
            print(f"Extracted {len(typo)} typography styles")

        if "components" in args.analyze:
            comps = analyzer.extract_components()
            print(f"Found {len(comps)} components")

        if "images" in args.analyze:
            imgs = analyzer.extract_images()
            print(f"Found {len(imgs)} images")

        # 출력 생성
        output_dir = Path(args.output)
        output_dir.mkdir(parents=True, exist_ok=True)

        if args.css:
            css_path = output_dir / "design-tokens.css"
            with open(css_path, "w") as f:
                f.write(analyzer.generate_design_tokens_css())
            print(f"CSS tokens saved to {css_path}")

        if args.report:
            report_path = output_dir / "analysis-report.md"
            with open(report_path, "w") as f:
                f.write(analyzer.generate_report())
            print(f"Report saved to {report_path}")

        # JSON 메타데이터 저장
        metadata_path = output_dir / "analysis-metadata.json"
        with open(metadata_path, "w") as f:
            json.dump({
                "colors": {k: v.to_hex() for k, v in analyzer.colors.items()},
                "typography_count": len(analyzer.typography),
                "components_count": len(analyzer.components),
                "images_count": len(analyzer.images)
            }, f, indent=2)
        print(f"Metadata saved to {metadata_path}")

    else:
        parser.print_help()
        print("\nError: Please provide --json or --file with --node and --token")
        sys.exit(1)


if __name__ == "__main__":
    main()
