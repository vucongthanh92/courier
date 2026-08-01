const notes = {
  focus: ["01 Focus Productivity", "Huong mac dinh can bang: web app chuyen nghiep, co sidebar, chat rong va drawer quan ly conversation."],
  zalo: ["02 Viet Social", "Huong gan voi nguoi dung Viet Nam: danh ba, QR, quick sticker, chat ca nhan/group de tiep can."],
  slack: ["03 Team Channels", "Huong team-work: server rail, channel list, thread panel va slash command de mo rong workflow."],
  stories: ["04 Social Stories", "Huong social/messenger: story row, reaction noi, bubble mem va sticker tray cho chat nhieu cam xuc."],
  command: ["05 Command Center", "Huong support/ops: queue, SLA metrics, ticket conversation, macro reply va customer profile."],
  zen: ["06 Zen Minimal", "Huong toi gian: mot khung chat yen tinh, it phan tan, phu hop direct message va mobile ca nhan."],
  mobile: ["07 Mobile Native", "Huong mobile-first: phone frame, bottom nav, target lon, conversation list va chat cung mot luong cham."],
  community: ["08 Community Rooms", "Huong cong dong: room map, announcement, slow mode, online members va phong chat dong nguoi."],
  inbox: ["09 Inbox Pro", "Huong email/inbox: bo loc uu tien, reading pane, tags, quick reply va review conversation nhu cong viec."],
  spatial: ["10 Spatial Glass", "Huong premium: panel noi, blur layer, context card va motion background cho demo cao cap."]
};

const title = document.querySelector("#note-title");
const copy = document.querySelector("#note-copy");
const mockups = document.querySelectorAll(".mockup");
const buttons = document.querySelectorAll(".switcher button");

buttons.forEach((button) => {
  button.addEventListener("click", () => {
    const target = button.dataset.target;
    buttons.forEach((item) => item.classList.toggle("active", item === button));
    mockups.forEach((mockup) => mockup.classList.toggle("active", mockup.classList.contains(target)));
    title.textContent = notes[target][0];
    copy.textContent = notes[target][1];
  });
});
