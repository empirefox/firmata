import 'package:flutter/material.dart';

import 'package:get/get.dart';
import 'package:group_list_view/group_list_view.dart';
import 'package:reactive_advanced_switch/reactive_advanced_switch.dart';
import 'package:reactive_forms/reactive_forms.dart';

import '../controllers/planet_edit_controller.dart';

class PlanetEditView extends StatelessWidget {
  static final intValueAccessor = IntValueAccessor();
  final controller = Get.find<PlanetEditController>();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          onPressed: Get.back,
          icon: Icon(Icons.chevron_left),
        ),
        title: Text('Edit Planet'),
        actions: [
          IconButton(
            onPressed: controller.about,
            icon: Icon(Icons.favorite),
          ),
        ],
      ),
      body: controller.invalid
          ? Center(
              child: Text(
                'Invalid planet to edit',
                style: TextStyle(fontSize: 20),
              ),
            )
          : _form(context),
    );
  }

  Widget _form(BuildContext context) {
    return ReactiveForm(
      formGroup: controller.form,
      child: Column(
        children: [
          _listView(context),
          ReactiveFormConsumer(
            builder: (context, form, child) {
              return ElevatedButton(
                child: Text('Submit'),
                onPressed: form.valid ? controller.onSubmit : null,
              );
            },
          ),
        ],
      ),
    );
  }

  Widget _listView(BuildContext context) {
    final fields = _fields(context);
    final headers = fields.keys.toList();
    final elements = fields.values.toList();
    return GroupListView(
      sectionsCount: headers.length,
      countOfItemInSection: (int section) => elements[section].length,
      separatorBuilder: (context, index) => const SizedBox(height: 5),
      sectionSeparatorBuilder: (context, section) => const SizedBox(height: 10),
      itemBuilder: (BuildContext context, IndexPath index) {
        return elements[index.section][index.index];
      },
      groupHeaderBuilder: (BuildContext context, int section) {
        return Padding(
          padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 8),
          child: Text(
            headers[section],
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
          ),
        );
      },
    );
  }

  Map<String, List<Widget>> _fields(BuildContext context) {
    return {
      'Target': [
        ReactiveTextField<String>(
          formControlName: 'name',
          textInputAction: TextInputAction.next,
          decoration: InputDecoration(
            labelText: 'Name',
          ),
        ),
        ReactiveTextField<String>(
          formControlName: 'host',
          textInputAction: TextInputAction.next,
          decoration: InputDecoration(
            labelText: 'Host',
          ),
        ),
        ReactiveTextField<int>(
          formControlName: 'port',
          decoration: InputDecoration(
            labelText: 'Port',
          ),
          valueAccessor: intValueAccessor,
        ),
        ReactiveTextField<String>(
          formControlName: 'userAgent',
          decoration: InputDecoration(
            labelText: 'User Agent',
          ),
        ),
      ],
      'Security': [
        ReactiveAdvancedSwitch<bool>(
          formControlName: 'isTlsDisabled',
          decoration: InputDecoration(
            labelText: 'Disable TLS',
          ),
        ),
        ReactiveTextField<String>(
          formControlName: 'tlsCertificates',
          decoration: InputDecoration(
            labelText: 'TLS Certificates',
          ),
          minLines: 5,
          maxLines: 10,
        ),
        ReactiveTextField<String>(
          formControlName: 'tlsPassword',
          decoration: InputDecoration(
            labelText: 'TLS Password',
          ),
        ),
        ReactiveTextField<String>(
          formControlName: 'tlsAuthority',
          decoration: InputDecoration(
            labelText: 'TLS Authority',
          ),
        ),
        ReactiveAdvancedSwitch<bool>(
          formControlName: 'canTlsInsecureSkipVerify',
          decoration: InputDecoration(
            labelText: 'TLS Insecure Skip Verify',
          ),
        ),
      ],
      'Codec': [
        ReactiveAdvancedSwitch<bool>(
          formControlName: 'supportGrpcCodecGzip',
          decoration: InputDecoration(
            labelText: 'gzip',
          ),
        ),
        ReactiveAdvancedSwitch<bool>(
          formControlName: 'supportGrpcCodecIdentity',
          decoration: InputDecoration(
            labelText: 'identity',
          ),
        ),
      ],
      'Auth': [
        ReactiveTextField<String>(
          formControlName: 'tokenType',
          decoration: InputDecoration(
            labelText: 'Token Type',
          ),
        ),
        ReactiveTextField<String>(
          formControlName: 'token',
          decoration: InputDecoration(
            labelText: 'Token',
          ),
        ),
      ],
      'Timeout': [
        ReactiveTextField<int>(
          formControlName: 'connectionTimeoutSeconds',
          decoration: InputDecoration(
            labelText: 'Connection Timeout',
            suffixText: 's',
          ),
          valueAccessor: intValueAccessor,
        ),
        ReactiveTextField<int>(
          formControlName: 'idleTimeoutSeconds',
          decoration: InputDecoration(
            labelText: 'Idle Timeout',
            suffixText: 's',
          ),
          valueAccessor: intValueAccessor,
        ),
        ReactiveTextField<int>(
          formControlName: 'callTimeoutSeconds',
          decoration: InputDecoration(
            labelText: 'Call Timeout',
            suffixText: 's',
          ),
          valueAccessor: intValueAccessor,
        ),
      ],
    };
  }
}
