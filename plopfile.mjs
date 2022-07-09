export default function (plop) {
  const today = new Date();
  const year = `${today.getFullYear()}`.padStart(4, "0");
  const month = `${today.getMonth() + 1}`.padStart(2, "0");

  plop.setHelper("getDate", function () {
    const date = today.getDate();
    const monthEn = [
      "Jan",
      "Feb",
      "Mar",
      "Apr",
      "May",
      "Jun",
      "Jul",
      "Aug",
      "Sep",
      "Oct",
      "Nov",
      "Dec",
    ][today.getMonth()];

    return `${date} ${monthEn} ${year}`;
  });

  plop.setGenerator("posts", {
    description: "new posts",
    prompts: [
      {
        type: "input",
        name: "slug",
        message: "the slug of the post",
      },
      {
        type: "input",
        name: "title",
        message: "the title of the post",
      },
    ],
    actions: [
      {
        type: "add",
        path: `src/pages/posts/${year}/${month}/{{slug}}.md`,
        templateFile: "templates/post.md",
      },
    ],
  });
}
