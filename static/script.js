const fileInput = document.getElementById("fileInput");
const textArea = document.querySelector("textArea");
const loadingPopup = document.getElementById("loading-popup");
const link = document.getElementById("link");

fileInput.addEventListener("change", (e) => {
  loadingPopup.style.display = "flex";
  const file = e.target.files[0];
  const formData = new FormData();
  console.log("file is: ", file);

  formData.append("image", file);
  formData.append("name", file["name"]);
  formData.append("type", file["type"]);

  fetch("/upload", {
    method: "POST",
    body: formData,
  })
    .then((res) => res.json())
    .then((data) => {
      loadingPopup.style.display = "none";
      textArea.value = data.content;
      link.href = data.path;
      link.download = data.filename;
    });
});
