import { generateStaticParamsFor, importPage } from "nextra/pages";
import { MDXWrapper } from "./mdx-wrapper";

export const generateStaticParams = generateStaticParamsFor("");

export async function generateMetadata(props: PageProps) {
  const params = await props.params;
  const mdxPath = params.mdxPath ?? [];
  const { metadata } = await importPage(mdxPath);
  return metadata;
}

type PageProps = {
  params: Promise<{ mdxPath?: string[] }>;
};

export default async function Page(props: PageProps) {
  const params = await props.params;
  const mdxPath = params.mdxPath ?? [];
  const result = await importPage(mdxPath);
  const { default: MDXContent, toc, metadata, sourceCode } = result;

  return (
    <MDXWrapper toc={toc} metadata={metadata} sourceCode={sourceCode}>
      <MDXContent {...props} params={params} />
    </MDXWrapper>
  );
}
