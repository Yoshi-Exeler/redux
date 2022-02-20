import { Component } from '@angular/core';
import { API, File, Folder } from '../api/api';

@Component({
  selector: 'app-tab1',
  templateUrl: 'tab1.page.html',
  styleUrls: ['tab1.page.scss']
})
export class Tab1Page {

  files: File[] = [];
  folders: Folder[] = [];
  path: string = "";


  constructor() {
    this.getData();
  }

  openFileContext(file: File) {
    console.log("Placeholder for file action", file);
  }

  canNavigateBack(): boolean {
    return this.path.length > 0
  }

  navigateIntoFolder(folder: Folder) {
    if (this.path.length === 0) {
      this.path = "/"
    }
    this.path += folder.Name + "/"
    console.log("Navigate to ", this.path);
    this.getData();
  }

  navigateBack() {
    let pathSegments = this.path.split("/");
    if (pathSegments.length === 3) {
      this.path = "";
      console.log("Navigate back to origin ", this.path)
      this.getData();
      return;
    }
    // navigate back to the last Segment
    let trunc = this.path.substring(0, this.path.lastIndexOf("/"));
    let withoutSegment = trunc.substring(0, trunc.lastIndexOf("/")) + "/";
    this.path = withoutSegment;
    console.log("Navigate back to path ", this.path)
    this.getData();
  }

  getData() {
    API.GetFolderContent(this.path).then((resp) => {
      this.files = resp.Files;
      this.folders = resp.Folders;
    })
  }

  getIconForType(type: string): string {
    // video types
    if (
      type === "mp4" ||
      type === "mov" ||
      type === "webm") {
      return "videocam-outline";
    }
    // text types
    if (
      type === "txt" ||
      type === "odt" ||
      type === "pdf") {
      return "document-text-outline";
    }
    return "document-outline";
  }

}

