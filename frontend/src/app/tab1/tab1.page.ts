import { Component } from '@angular/core';
import axios from 'axios'
import { API, File, Folder } from '../api/api';

@Component({
  selector: 'app-tab1',
  templateUrl: 'tab1.page.html',
  styleUrls: ['tab1.page.scss']
})
export class Tab1Page {

  mockFiles: File[] = []
  mockFolders: Folder[] = []

  constructor() {
    this.getData();
  }

  getData() {
    API.GetFolderContent("/arbeit").then((resp) => {
      this.mockFiles = resp.Files;
      this.mockFolders = resp.Folders;
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

