import { Component } from "@angular/core";

@Component({
  selector: "app-tabs",
  templateUrl: "tabs.page.html",
  styleUrls: ["tabs.page.scss"],
})
export class TabsPage {
  constructor() {}

  getCurrentPage(): string {
    return window.location.href.substring(
      window.location.href.lastIndexOf("/") + 1
    );
  }
}
