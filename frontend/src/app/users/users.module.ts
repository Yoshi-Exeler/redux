import { IonicModule } from '@ionic/angular';
import { RouterModule } from '@angular/router';
import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormsModule, ReactiveFormsModule } from '@angular/forms';
import { UsersPage } from './users.page';

import { UsersPageRoutingModule } from './users-routing.module';

@NgModule({
  imports: [
    IonicModule,
    CommonModule,
    FormsModule,
    RouterModule.forChild([{ path: '', component: UsersPage }]),
    UsersPageRoutingModule,
    FormsModule,
    ReactiveFormsModule
  ],
  declarations: [UsersPage]
})
export class UsersPageModule { }
