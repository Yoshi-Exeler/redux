import { Component, Input, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from "@angular/forms";

@Component({
  selector: 'app-user-add-modal',
  templateUrl: './user-add-modal.component.html',
  styleUrls: ['./user-add-modal.component.scss'],
})
export class UserAddModalComponent implements OnInit {

  @Input() reference: any;

  fgUser: FormGroup;

  constructor(private formBuilder: FormBuilder) { }

  ngOnInit(): void {
    this.fgUser = this.formBuilder.group({
      username: [null, []],
      password: [null, []],
      admin: [null, []],
    });
  }

  close() {
    this.reference.modalController.dismiss();
  }

  addUser() {
    this.reference.addUser(this.fgUser.value.username,this.fgUser.value.password,this.fgUser.value.admin);
    this.reference.modalController.dismiss();
  }

}
