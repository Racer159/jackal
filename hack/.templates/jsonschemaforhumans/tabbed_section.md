<blockquote>

{{ current_node | md_array_items(title) | md_generate_table }}

{% for node in current_node.array_items %}
<blockquote>

    {% filter md_heading(2, node.html_id) -%}
        Property `{% with schema=node %}{%- include "breadcrumbs.md" %}{% endwith %}`
    {%- endfilter %}

    {% with schema=node, skip_headers=False %}
        {% include "content.md" %}
    {% endwith %}

</blockquote>
{% endfor %}

</blockquote>
