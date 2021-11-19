import 'package:flutter/material.dart';

import 'package:get/get.dart';

import '../controllers/planets_controller.dart';

class PlanetsView extends StatelessWidget {
  final controller = Get.find<PlanetsController>();
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          onPressed: Get.back,
          icon: Icon(Icons.chevron_left),
        ),
        title: Text('Planets'),
        actions: [
          IconButton(
            onPressed: controller.create,
            icon: Icon(Icons.add_circle),
          ),
          IconButton(
            onPressed: controller.about,
            icon: Icon(Icons.favorite),
          ),
        ],
      ),
      body: GetBuilder(
        id: 'planets',
        init: controller,
        builder: (_) => ListView.builder(
          itemCount: controller.planets.length,
          itemBuilder: _itemBuilder,
        ),
      ),
    );
  }

  Widget _itemBuilder(BuildContext context, int index) {
    final planet = controller.planets[index];
    return GetBuilder(
      id: planet.id,
      init: controller,
      builder: (_) {
        return ListTile(
          title: FutureBuilder(
            future: controller.findServerName(planet.id),
            builder: (BuildContext context, AsyncSnapshot<String> snapshot) =>
                Text(planet.viewName(snapshot.data)),
          ),
          subtitle: Text(planet.address),
          trailing: Row(
            children: [
              IconButton(
                onPressed: () => controller.delete(planet),
                icon: Icon(
                  Icons.cancel,
                  color: Colors.red,
                ),
              ),
              IconButton(
                onPressed: () => controller.edit(planet),
                icon: Icon(
                  Icons.edit,
                  color: Colors.green,
                ),
              ),
              IconButton(
                onPressed: controller.isOnline(planet)
                    ? () => controller.shutdown(planet)
                    : null,
                icon: Icon(
                  Icons.cancel,
                  color: Colors.blue,
                ),
              ),
              IconButton(
                onPressed: () => controller.view(planet),
                icon: Icon(
                  Icons.cancel,
                  color: Colors.purple,
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
