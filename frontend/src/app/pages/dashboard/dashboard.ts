import { Component, OnInit } from '@angular/core';
import { Router, RouterLink, RouterLinkActive } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { MatIconModule } from '@angular/material/icon';
import { Auth } from '../../core/services/auth';

@Component({
  selector: 'app-dashboard',
  imports: [MatToolbarModule, MatButtonModule, MatCardModule, MatSidenavModule, MatListModule, MatIconModule, RouterLink, RouterLinkActive],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss',
})
export class Dashboard implements OnInit {
  stats = {
    servers: 0,
    libraries: 0,
    devices: 0,
    sessions: 0
  };

  constructor(private authService: Auth, private router: Router) {}

  ngOnInit(): void {
    // TODO: 从后端加载统计数据
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
