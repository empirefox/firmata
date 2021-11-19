import 'package:flutter/material.dart';

import 'package:get/get.dart';

import '../controllers/home_controller.dart';

class HomeView extends StatelessWidget {
  final controller = Get.find<HomeController>();
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('HomeView'),
        centerTitle: true,
      ),
      body: Center(
        child: Text(
          'HomeView is working',
          style: TextStyle(fontSize: 20),
        ),
      ),
    );
  }
}
